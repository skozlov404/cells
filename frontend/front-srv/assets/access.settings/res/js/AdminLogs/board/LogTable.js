/*
 * Copyright 2007-2020 Charles du Jeu - Abstrium SAS <team (at) pyd.io>
 * This file is part of Pydio.
 *
 * Pydio is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

import React from 'react'
import Pydio from 'pydio'
import {Paper, FontIcon, IconButton} from 'material-ui'
const {MaterialTable} = Pydio.requireLib('components');
import Log from '../model/Log'
import LogDetail from './LogDetail'
const {moment} = Pydio.requireLib('boot');

class LogTable extends React.Component {

    constructor(props){
        super(props);
        this.state = {logs: [], loading: false, rootSpans:{}, selectedRows: []};
    }

    initRootSpans(logs){
        if(!logs || !logs.length) {
            return {logs: [], rootSpans:{}};
        }
        const rootSpans = {};
        // Detect logs without parent Uuid
        let oLogs = logs.map((l, i) => {
            if(!l.SpanUuid){
                l.SpanUuid = 'span-' + i;
            }
            if (!l.SpanRootUuid) {
                rootSpans[l.SpanUuid] = {open: true, children: []};
            }
            return l;
        });
        // Filter out logs with parent Uuid and place them as children
        let result = [];
        for (let i = 0; i< oLogs.length; i++){
            let l = oLogs[i];
            if(l.SpanRootUuid && rootSpans[l.SpanRootUuid]){
                rootSpans[l.SpanRootUuid].children.push(l);
                l.HasRoot = true;
                continue;
            } else if(l.SpanRootUuid && !rootSpans[l.SpanRootUuid]){
                // Create a fake root
                let root = {...l};
                root.SpanUuid = l.SpanRootUuid;
                l.HasRoot = true;
                rootSpans[l.SpanRootUuid] = {open:true, children: [l]};
                result.push(root);
                continue;
            }
            result.push(l);
        }
        return {logs: result, rootSpans};
    }

    /**
     * @return []{*}
     */
    openSpans(){
        const {logs, rootSpans} = this.state;
        let result = [];
        for (let j = 0; j < logs.length; j++){
            let l = logs[j];
            let root = rootSpans[l.SpanUuid];
            if(root.children.length){
                l.HasChildren = true;
                l.IsOpen = root.open;
                if(!l.RemoteAddress) {
                    const cRemote = root.children.filter(c => c.RemoteAddress);
                    if(cRemote.length) l.RemoteAddress = cRemote[0].RemoteAddress;
                }
            }
            result.push(l);
            if(root.open){
                result = [...result, ...root.children];
            }
        }
        return result;
    }

    load(service, query, page, size, contentType, onLoadingStatusChange){

        const {logs} = this.state;

        if(onLoadingStatusChange) {
            this.setState({loading: true});
            onLoadingStatusChange(true, logs.length);
        }
        Pydio.startLoading();
        Log.loadLogs(service, query, page, size, contentType).then((data) => {
            Pydio.endLoading();
            const {logs, rootSpans} = this.initRootSpans(data.Logs);
            this.setState({logs, rootSpans, loading: false}, () => {
                if(onLoadingStatusChange) {
                    onLoadingStatusChange(false, (data.Logs?data.Logs.length:0));
                }
            });
        }).catch(reason => {
            Pydio.endLoading();
            if(onLoadingStatusChange) {
                this.setState({loading: false});
                onLoadingStatusChange(false, logs.length);
            }
        });

    }

    componentWillMount(){
        const {service, page, size, onLoadingStatusChange} = this.props;
        this.load(service, '', page, size, 'JSON', onLoadingStatusChange);
    }

    componentWillReceiveProps(nextProps) {

        const {service, query, page, size, onLoadingStatusChange, z} = nextProps;
        if(query === this.props.query && size === this.props.size && page === this.props.page && z === this.props.z){
            return;
        }
        this.load(service, query, page, size, 'JSON', onLoadingStatusChange);
    }

    render(){
        const {loading, rootSpans, selectedRows} = this.state;
        const {pydio, onTimestampContext, query, focus} = this.props;
        const {onPageNext, onPagePrev, nextDisabled, prevDisabled, onPageSizeChange, page, size, pageSizes} = this.props;
        let logs = this.openSpans();
        if(selectedRows.length){
            const expStyle = {paddingBottom: 20, paddingLeft: 53, backgroundColor: '#fafafa', marginTop: -10, paddingTop: 10};
            const first = JSON.stringify(selectedRows[0]);
            logs = logs.map(log => {
                if(JSON.stringify(log) === first){
                    return {
                        ...log,
                        expandedRow:(
                            <LogDetail
                                style={expStyle}
                                userDisplay={"inline"}
                                pydio={pydio}
                                log={log}
                                focus={focus}
                                onSelectPeriod={onTimestampContext}
                                onRequestClose={()=> this.setState({selectedRows:[]})}
                            />
                        )}
                } else {
                    return log;
                }
            })
        }
        const {MessageHash} = pydio;

        const columns = [
            {
                name:'Root',
                label:'',
                style:{width: 20, paddingLeft:0,paddingRight:0, overflow:'visible'},
                headerStyle:{width:20, paddingLeft:0,paddingRight:0},
                renderCell:(row) => {
                    if(row.HasChildren){
                        const toggle = () => {
                            rootSpans[row.SpanUuid].open = !rootSpans[row.SpanUuid].open;
                            this.setState({rootSpans});
                        };
                        return <IconButton
                            iconClassName={row.IsOpen?"mdi mdi-menu-down":"mdi mdi-menu-right"}
                            onTouchTap={toggle}
                            onClick={e =>e.stopPropagation()}
                        />
                    }
                    return null;
                }
            },
            {name:'Ts', label: pydio.MessageHash["settings.17"], renderCell:(row)=>{
                const m = moment(row.Ts * 1000);
                let dateString;
                if (m.isSame(Date.now(), 'day')){
                    dateString = m.format('HH:mm:ss');
                } else {
                    dateString = m.toLocaleString();
                }
                if(row.HasRoot){
                    return <span style={{display:'flex', alignItems:'center'}}><FontIcon className={"mdi mdi-play-circle-outline"} style={{fontSize: 12, marginRight: 5}}/> {dateString}</span>
                }
                return dateString;
            }, style:{width: 130, padding: 12}, headerStyle:{width: 130, padding: 12}},
            {name:'Logger', label:MessageHash['ajxp_admin.logs.service'], hideSmall:true, renderCell:(row) => {return row['Logger'] ? row['Logger'].replace('pydio.', '') : ''}, style:{width: 130, padding: '12px 0'}, headerStyle:{width: 130, padding: '12px 0'}},
            {name:'Msg', label:MessageHash['ajxp_admin.logs.message'], renderCell:(row)=>{
                let msg = row.Msg;
                if(row.NodePath){
                    msg += ` [${row.NodePath}]`;
                } else if(row.NodeUuid){
                    msg += ` [${row.NodeUuid}]`;
                }
                return msg;
            }},
        ];

        const {body} = AdminComponents.AdminStyles();
        const {tableMaster} = body;
        let pagination;
        if(onPageNext){
            pagination = {
                page:(page + 1),
                pageSize: size,
                pageSizes,
                onPageNext:v => onPageNext(v -1 ),
                onPagePrev:v => onPagePrev(v -1),
                onPageSizeChange,
                nextDisabled,
                prevDisabled
            };
        }

        return (
            <MaterialTable
                data={logs}
                columns={columns}
                onSelectRows={(rows) => {
                    this.setState({selectedRows: rows});
                    if(this.props.onTimestampContext){
                        this.props.onTimestampContext(null);
                    }
                }}
                deselectOnClickAway={true}
                showCheckboxes={false}
                emptyStateString={loading ? MessageHash['settings.33']: (query) ? MessageHash['ajxp_admin.logs.noresults'] : MessageHash['ajxp_admin.logs.noentries']}
                computeRowStyle={(row) => {
                    let style = {};
                    if (row.HasRoot){
                        style.backgroundColor = '#F5F5F5';
                    }
                    if (row.Level === 'error') {
                        style.color = '#E53935';
                    }
                    return style;
                }}
                masterStyles={tableMaster}
                pagination={pagination}
            />
        );
    }
}

export {LogTable as default}
