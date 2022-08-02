package resource

import (
	"context"
)

// Resource represents a resource instance. This is the core interface that all
// resources must implement.
//
// It is also conventional for resources to implement the
// ResourceWithImportState interface, which enables practitioners to import
// existing infrastructure into Terraform.
type Resource interface {
	// Create is called when the provider must create a new resource. Config
	// and planned state values should be read from the
	// CreateRequest and new state values set on the CreateResponse.
	Create(context.Context, CreateRequest, *CreateResponse)

	// Read is called when the provider must read resource values in order
	// to update state. Planned state values should be read from the
	// ReadRequest and new state values set on the ReadResponse.
	Read(context.Context, ReadRequest, *ReadResponse)

	// Update is called to update the state of the resource. Config, planned
	// state, and prior state values should be read from the
	// UpdateRequest and new state values set on the UpdateResponse.
	Update(context.Context, UpdateRequest, *UpdateResponse)

	// Delete is called when the provider must delete the resource. Config
	// values may be read from the DeleteRequest.
	//
	// If execution completes without error, the framework will automatically
	// call DeleteResponse.State.RemoveResource(), so it can be omitted
	// from provider logic.
	Delete(context.Context, DeleteRequest, *DeleteResponse)
}

// ResourceWithConfigValidators is an interface type that extends Resource to include declarative validations.
//
// Declaring validation using this methodology simplifies implmentation of
// reusable functionality. These also include descriptions, which can be used
// for automating documentation.
//
// Validation will include ConfigValidators and ValidateConfig, if both are
// implemented, in addition to any Attribute or Type validation.
type ResourceWithConfigValidators interface {
	Resource

	// ConfigValidators returns a list of functions which will all be performed during validation.
	ConfigValidators(context.Context) []ConfigValidator
}

// Optional interface on top of Resource that enables provider control over
// the ImportResourceState RPC. This RPC is called by Terraform when the
// `terraform import` command is executed. Afterwards, the ReadResource RPC
// is executed to allow providers to fully populate the resource state.
type ResourceWithImportState interface {
	Resource

	// ImportState is called when the provider must import the state of a
	// resource instance. This method must return enough state so the Read
	// method can properly refresh the full resource.
	//
	// If setting an attribute with the import identifier, it is recommended
	// to use the ImportStatePassthroughID() call in this method.
	ImportState(context.Context, ImportStateRequest, *ImportStateResponse)
}

// ResourceWithModifyPlan represents a resource instance with a ModifyPlan
// function.
type ResourceWithModifyPlan interface {
	Resource

	// ModifyPlan is called when the provider has an opportunity to modify
	// the plan: once during the plan phase when Terraform is determining
	// the diff that should be shown to the user for approval, and once
	// during the apply phase with any unknown values from configuration
	// filled in with their final values.
	//
	// The planned new state is represented by
	// ModifyPlanResponse.Plan. It must meet the following
	// constraints:
	// 1. Any non-Computed attribute set in config must preserve the exact
	// config value or return the corresponding attribute value from the
	// prior state (ModifyPlanRequest.State).
	// 2. Any attribute with a known value must not have its value changed
	// in subsequent calls to ModifyPlan or Create/Read/Update.
	// 3. Any attribute with an unknown value may either remain unknown
	// or take on any value of the expected type.
	//
	// Any errors will prevent further resource-level plan modifications.
	ModifyPlan(context.Context, ModifyPlanRequest, *ModifyPlanResponse)
}

// Optional interface on top of Resource that enables provider control over
// the UpgradeResourceState RPC. This RPC is automatically called by Terraform
// when the current Schema type Version field is greater than the stored state.
// Terraform does not store previous Schema information, so any breaking
// changes to state data types must be handled by providers.
//
// Terraform CLI can execute the UpgradeResourceState RPC even when the prior
// state version matches the current schema version. The framework will
// automatically intercept this request and attempt to respond with the
// existing state. In this situation the framework will not execute any
// provider defined logic, so declaring it for this version is extraneous.
type ResourceWithUpgradeState interface {
	Resource

	// A mapping of prior state version to current schema version state upgrade
	// implementations. Only the specified state upgrader for the prior state
	// version is called, rather than each version in between, so it must
	// encapsulate all logic to convert the prior state to the current schema
	// version.
	//
	// Version keys begin at 0, which is the default schema version when
	// undefined. The framework will return an error diagnostic should the
	// requested state version not be implemented.
	UpgradeState(context.Context) map[int64]StateUpgrader
}

// ResourceWithValidateConfig is an interface type that extends Resource to include imperative validation.
//
// Declaring validation using this methodology simplifies one-off
// functionality that typically applies to a single resource. Any documentation
// of this functionality must be manually added into schema descriptions.
//
// Validation will include ConfigValidators and ValidateConfig, if both are
// implemented, in addition to any Attribute or Type validation.
type ResourceWithValidateConfig interface {
	Resource

	// ValidateConfig performs the validation.
	ValidateConfig(context.Context, ValidateConfigRequest, *ValidateConfigResponse)
}