/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package grpc

import (
	"context"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/metadata"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/pydio/cells/broker/activity"
	"github.com/pydio/cells/common"
	"github.com/pydio/cells/common/auth"
	"github.com/pydio/cells/common/log"
	activity2 "github.com/pydio/cells/common/proto/activity"
	"github.com/pydio/cells/common/proto/idm"
	"github.com/pydio/cells/common/proto/tree"
	"github.com/pydio/cells/common/registry"
	"github.com/pydio/cells/common/service/context"
	"github.com/pydio/cells/common/service/proto"
	context2 "github.com/pydio/cells/common/utils/context"
	"github.com/pydio/cells/common/utils/permissions"
	"github.com/pydio/cells/common/views"
)

type MicroEventsSubscriber struct {
	sync.Mutex
	treeClient tree.NodeProviderClient
	usrClient  idm.UserServiceClient
	roleClient idm.RoleServiceClient
	wsClient   idm.WorkspaceServiceClient

	parentsCache     *cache.Cache
	accessListsCache *cache.Cache
	changeEvents     []*idm.ChangeEvent
	aclsChan         chan *idm.ChangeEvent
	dao              activity.DAO
}

func NewEventsSubscriber(dao activity.DAO) *MicroEventsSubscriber {
	m := &MicroEventsSubscriber{
		dao:              dao,
		aclsChan:         make(chan *idm.ChangeEvent),
		parentsCache:     cache.New(3*time.Minute, 10*time.Minute),
		accessListsCache: cache.New(20*time.Second, 60*time.Second),
	}
	go m.DebounceAclsEvents()
	return m
}

func (e *MicroEventsSubscriber) getTreeClient() tree.NodeProviderClient {
	if e.treeClient == nil {
		e.treeClient = tree.NewNodeProviderClient(registry.GetClient(common.ServiceTree))
	}
	return e.treeClient
}

func (e *MicroEventsSubscriber) getUserClient() idm.UserServiceClient {
	if e.usrClient == nil {
		e.usrClient = idm.NewUserServiceClient(registry.GetClient(common.ServiceUser))
	}
	return e.usrClient
}

func (e *MicroEventsSubscriber) getRoleClient() idm.RoleServiceClient {
	if e.roleClient == nil {
		e.roleClient = idm.NewRoleServiceClient(registry.GetClient(common.ServiceRole))
	}
	return e.roleClient
}

func (e *MicroEventsSubscriber) getWorkspaceClient() idm.WorkspaceServiceClient {
	if e.wsClient == nil {
		e.wsClient = idm.NewWorkspaceServiceClient(registry.GetClient(common.ServiceWorkspace))
	}
	return e.wsClient
}

// Handle processes the received events and sends them to the subscriber
func (e *MicroEventsSubscriber) HandleNodeChange(ctx context.Context, msg *tree.NodeChangeEvent) error {

	dao := servicecontext.GetDAO(ctx).(activity.DAO)

	author := common.PydioSystemUsername
	meta, ok := metadata.FromContext(ctx)
	if ok {
		user, exists := meta[common.PydioContextUserKey]
		user1, exists1 := meta[strings.ToLower(common.PydioContextUserKey)]
		if exists {
			author = user
		} else if exists1 {
			author = user1
		}
	}
	if author == common.PydioSystemUsername {
		// Ignore events triggered by initial sync
		return nil
	}
	log.Logger(ctx).Debug("Fan out event to activities", zap.String(common.KEY_USER, author), msg.Zap())

	// Create Activities and post them to associated inboxes
	ac, Node := activity.DocumentActivity(author, msg)
	if Node != nil && Node.Uuid != "" {

		loadedNode, parentUuids := e.parentsFromCache(ctx, Node, msg.Type == tree.NodeChangeEvent_DELETE)
		// Use reloaded node
		if msg.Type == tree.NodeChangeEvent_UPDATE_USER_META {
			if Node.MetaStore != nil {
				if loadedNode.MetaStore == nil {
					loadedNode.MetaStore = make(map[string]string, len(Node.MetaStore))
				}
				for k, v := range Node.MetaStore {
					loadedNode.MetaStore[k] = v
				}
			}
			msg.Target = loadedNode
			// Rebuild activity
			ac, Node = activity.DocumentActivity(author, msg)
		}

		// Ignore hidden files
		if tree.IgnoreNodeForOutput(ctx, Node) {
			return nil
		}

		//
		// Post to the initial node Outbox
		//
		log.Logger(ctx).Debug("Posting Activity to node outbox")
		if msg.Type == tree.NodeChangeEvent_DELETE {
			dao.Delete(activity2.OwnerType_NODE, Node.Uuid)
		} else {
			dao.PostActivity(activity2.OwnerType_NODE, Node.Uuid, activity.BoxOutbox, ac, nil)
		}

		//
		// Post to the author Outbox
		//
		log.Logger(ctx).Debug("Posting Activity to author outbox")
		dao.PostActivity(activity2.OwnerType_USER, author, activity.BoxOutbox, ac, ctx)

		//
		// Post to parents Outbox'es as well
		//
		for _, uuid := range parentUuids {
			dao.PostActivity(activity2.OwnerType_NODE, uuid, activity.BoxOutbox, ac, nil)
		}

		//
		// Find followers and post activity to their Inbox
		//
		subUuids := parentUuids
		if msg.Type != tree.NodeChangeEvent_CREATE {
			subUuids = append(subUuids, Node.Uuid)
		}
		subscriptions, err := dao.ListSubscriptions(activity2.OwnerType_NODE, subUuids)
		log.Logger(ctx).Debug("Listing followers on node and its parents", zap.Any("subs", subscriptions))
		if err != nil {
			return err
		}
		for _, subscription := range subscriptions {

			if len(subscription.Events) == 0 {
				continue
			}
			// Ignore if author is user
			if subscription.UserId == author {
				continue
			}
			evread := false
			evchange := false
			for _, ev := range subscription.Events {
				if strings.Compare(ev, "read") == 0 {
					evread = true
				} else if strings.Compare(ev, "change") == 0 {
					evchange = true
				}
			}
			if !((evread && ac.Type == activity2.ObjectType_Read) || (evchange && ac.Type != activity2.ObjectType_Read)) {
				continue
			}
			accessList, user, er := permissions.AccessListFromUser(ctx, subscription.UserId, false)
			if er != nil {
				log.Logger(ctx).Error("Could not load access list", zap.Error(er))
				continue
			}
			userCtx := auth.WithImpersonate(ctx, user)
			ancestors, ez := views.BuildAncestorsListOrParent(userCtx, e.getTreeClient(), loadedNode)
			if ez != nil {
				log.Logger(ctx).Error("Could not load ancestors list", zap.Error(er))
				continue
			}
			if accessList.CanReadWithResolver(userCtx, e.vNodeResolver, ancestors...) {
				dao.PostActivity(activity2.OwnerType_USER, subscription.UserId, activity.BoxInbox, ac, ctx)
			}
		}

	}

	return nil
}

func (e *MicroEventsSubscriber) vNodeResolver(ctx context.Context, n *tree.Node) (*tree.Node, bool) {
	pool := views.NewClientsPool(false)
	return views.GetVirtualNodesManager().GetResolver(pool, false)(ctx, n)
}

func (e *MicroEventsSubscriber) HandleIdmChange(ctx context.Context, msg *idm.ChangeEvent) error {

	if msg.Acl != nil {
		author, _ := permissions.FindUserNameInContext(ctx)
		if msg.Attributes == nil {
			msg.Attributes = make(map[string]string)
		}
		msg.Attributes["user"] = author
		e.aclsChan <- msg
	} else if msg.User != nil && msg.Type == idm.ChangeEventType_DELETE && msg.User.Login != "" {
		// Clear activity for deleted user
		ctx = servicecontext.WithServiceName(ctx, Name)
		ctx = servicecontext.WithServiceColor(ctx, servicecontext.ServiceColorGrpc)
		log.Logger(ctx).Debug("Clearing activities for user", msg.User.ZapLogin())
		go e.dao.Delete(activity2.OwnerType_USER, msg.User.Login)
	}

	return nil
}

func (e *MicroEventsSubscriber) parentsFromCache(ctx context.Context, node *tree.Node, isDel bool) (*tree.Node, []string) {

	e.Lock()
	defer e.Unlock()

	var parentUuids []string
	loadedNode := node

	// Current Node
	// parentUuids = append(parentUuids, node.Uuid)
	if node.Path == "" && !isDel {
		// Reload by Uuid
		if resp, err := e.getTreeClient().ReadNode(ctx, &tree.ReadNodeRequest{Node: &tree.Node{Uuid: node.Uuid}}); err == nil && resp.Node != nil {
			loadedNode = resp.Node
		}
	}
	// Manually load parents from Path
	parentPath := loadedNode.Path
	for {
		parentPath = path.Dir(parentPath)
		if parentPath == "" || parentPath == "/" || parentPath == "." {
			break
		}
		if pU, ok := e.parentsCache.Get(parentPath); ok {
			val := pU.(string)
			if val != "**DELETED**" {
				parentUuids = append(parentUuids, pU.(string))
			}
		} else {
			resp, err := e.getTreeClient().ReadNode(ctx, &tree.ReadNodeRequest{Node: &tree.Node{Path: parentPath}})
			if err == nil {
				uuid := resp.Node.Uuid
				e.parentsCache.Set(parentPath, uuid, cache.DefaultExpiration)
				parentUuids = append(parentUuids, uuid)
			} else if errors.Parse(err.Error()).Code == 404 {
				e.parentsCache.Set(parentPath, "**DELETED**", cache.DefaultExpiration)
			}
		}
	}

	return loadedNode, parentUuids
}

func (e *MicroEventsSubscriber) DebounceAclsEvents() {
	for {
		select {
		case acl := <-e.aclsChan:
			e.changeEvents = append(e.changeEvents, acl)
		case <-time.After(3 * time.Second):
			e.ProcessBuffer(e.changeEvents...)
			e.changeEvents = []*idm.ChangeEvent{}
		}
	}
}

func (e *MicroEventsSubscriber) ProcessBuffer(cE ...*idm.ChangeEvent) {
	if len(cE) == 0 {
		return
	}
	remaining := make(map[string]*idm.ChangeEvent)
	roles := make(map[string]*idm.Role)
	users := make(map[string]*idm.User)
	workspaces := make(map[string]*idm.Workspace)
	// Remove updates/deletes
	for _, event := range cE {
		if event.Type == idm.ChangeEventType_UPDATE && (event.Acl.Action.Name == permissions.AclRead.Name || event.Acl.Action.Name == permissions.AclWrite.Name) {
			var del bool
			for _, e2 := range cE {
				if e2.Type == idm.ChangeEventType_DELETE && e2.Acl.RoleID == event.Acl.RoleID && e2.Acl.WorkspaceID == event.Acl.WorkspaceID &&
					e2.Acl.Action.Name == event.Acl.Action.Name && e2.Acl.Action.Value == event.Acl.Action.Value {
					del = true
				}
			}
			if !del {
				key := event.Acl.RoleID + "-" + event.Acl.WorkspaceID
				if already, ok := remaining[key]; ok {
					already.Acl.Action.Name = already.Acl.Action.Name + "," + event.Acl.Action.Name
				} else {
					remaining[key] = event
				}
				roles[event.Acl.RoleID] = &idm.Role{}
				users[event.Attributes["user"]] = &idm.User{}
				workspaces[event.Acl.WorkspaceID] = &idm.Workspace{}
			}
		}
	}
	if len(remaining) == 0 || len(roles) == 0 || len(users) == 0 || len(workspaces) == 0 {
		return
	}

	// Load resources
	ctx := context2.WithUserNameMetadata(context.Background(), common.PydioSystemUsername)
	if er := e.LoadResources(ctx, roles, users, workspaces); er != nil {
		return
	}

	// For the moment, only handle real users (no roles or groups), and skip source == target
	for _, r := range remaining {
		sourceUser := users[r.Attributes["user"]]
		targetRole := roles[r.Acl.RoleID]
		if targetRole.Uuid == "" || !targetRole.UserRole || targetRole.Uuid == sourceUser.Uuid {
			//log.Logger(context.Background()).Info("Ignoring event", zap.Any("e", r))
			continue
		}
		if targetUser, ok := users[targetRole.Uuid]; ok {
			ac := activity.AclActivity(sourceUser.Login, workspaces[r.Acl.WorkspaceID], r.Acl.Action.Name)
			log.Logger(context.Background()).Debug("Publishing Activity", zap.String("targetUser", targetUser.Login), zap.Any("a", ac))
			// Post to target User Inbox
			e.dao.PostActivity(activity2.OwnerType_USER, targetUser.Login, activity.BoxInbox, ac, ctx)
			if workspaces[r.Acl.WorkspaceID].Scope == idm.WorkspaceScope_ROOM {
				// Post to node Outbox
				e.dao.PostActivity(activity2.OwnerType_NODE, r.Acl.NodeID, activity.BoxOutbox, ac, nil)
			}
		}
	}

}

func (e *MicroEventsSubscriber) LoadResources(ctx context.Context, roles map[string]*idm.Role, users map[string]*idm.User, workspaces map[string]*idm.Workspace) error {

	// Load Roles
	var rUuids []string
	for k, _ := range roles {
		rUuids = append(rUuids, k)
	}
	q, _ := ptypes.MarshalAny(&idm.RoleSingleQuery{Uuid: rUuids})
	streamer, er := e.getRoleClient().SearchRole(ctx, &idm.SearchRoleRequest{Query: &service.Query{SubQueries: []*any.Any{q}}})
	if er != nil {
		return er
	}
	defer streamer.Close()
	for {
		resp, err := streamer.Recv()
		if err != nil {
			break
		}
		roles[resp.Role.Uuid] = resp.Role
	}

	// Load Users (from logins and from roleIds)
	var queries []*any.Any
	for k, _ := range users {
		q, _ := ptypes.MarshalAny(&idm.UserSingleQuery{Login: k})
		queries = append(queries, q)
	}
	for _, role := range roles {
		if role.UserRole {
			q, _ := ptypes.MarshalAny(&idm.UserSingleQuery{Uuid: role.Uuid})
			queries = append(queries, q)
		}
	}
	stream2, er := e.getUserClient().SearchUser(ctx, &idm.SearchUserRequest{Query: &service.Query{
		SubQueries: queries,
		Operation:  service.OperationType_OR,
	}})
	if er != nil {
		return er
	}
	defer stream2.Close()
	for {
		resp, err := stream2.Recv()
		if err != nil {
			break
		}
		if _, o := users[resp.User.Login]; o {
			users[resp.User.Login] = resp.User
		} else {
			users[resp.User.Uuid] = resp.User
		}
	}

	// Load Workspaces
	var queriesW []*any.Any
	for k, _ := range workspaces {
		q, _ := ptypes.MarshalAny(&idm.WorkspaceSingleQuery{Uuid: k})
		queriesW = append(queriesW, q)
	}
	stream3, er := e.getWorkspaceClient().SearchWorkspace(ctx, &idm.SearchWorkspaceRequest{Query: &service.Query{SubQueries: queriesW}})
	if er != nil {
		return er
	}
	defer stream3.Close()
	for {
		resp, err := stream3.Recv()
		if err != nil {
			break
		}
		workspaces[resp.Workspace.UUID] = resp.Workspace
	}

	log.Logger(context.Background()).Debug("Loaded Resources", zap.Any("roles", roles), zap.Any("users", users), zap.Any("workspaces", workspaces))

	return nil
}
