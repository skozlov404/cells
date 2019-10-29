// Code generated by go-swagger; DO NOT EDIT.

package provisioning

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/pydio/pydio-sdk-go/models"
)

// AdminUpdateWorkspaceReader is a Reader for the AdminUpdateWorkspace structure.
type AdminUpdateWorkspaceReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *AdminUpdateWorkspaceReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewAdminUpdateWorkspaceOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewAdminUpdateWorkspaceOK creates a AdminUpdateWorkspaceOK with default headers values
func NewAdminUpdateWorkspaceOK() *AdminUpdateWorkspaceOK {
	return &AdminUpdateWorkspaceOK{}
}

/*AdminUpdateWorkspaceOK handles this case with default header values.

Workspace object
*/
type AdminUpdateWorkspaceOK struct {
	Payload *models.AdminWorkspace
}

func (o *AdminUpdateWorkspaceOK) Error() string {
	return fmt.Sprintf("[PATCH /admin/workspaces][%d] adminUpdateWorkspaceOK  %+v", 200, o.Payload)
}

func (o *AdminUpdateWorkspaceOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AdminWorkspace)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}