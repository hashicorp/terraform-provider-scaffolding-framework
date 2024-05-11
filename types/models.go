package types

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

/**
 * Secret models.
 */

type SecretModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectID      types.String `tfsdk:"project_id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
	Note           types.String `tfsdk:"note"`
	CreationDate   types.String `tfsdk:"creation_date"`
	RevisionDate   types.String `tfsdk:"revision_date"`
}

type JSONSecret struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	ProjectID      string `json:"projectId"`
	Key            string `json:"key"`
	Value          string `json:"value"`
	Note           string `json:"note"`
	CreationDate   string `json:"creationDate"`
	RevisionDate   string `json:"revisionDate"`
}

func (j *JSONSecret) Parse() SecretModel {
	return SecretModel{
		ID:             types.StringValue(j.ID),
		OrganizationID: types.StringValue(j.OrganizationID),
		ProjectID:      types.StringValue(j.ProjectID),
		Key:            types.StringValue(j.Key),
		Value:          types.StringValue(j.Value),
		Note:           types.StringValue(j.Note),
		CreationDate:   types.StringValue(j.CreationDate),
		RevisionDate:   types.StringValue(j.RevisionDate),
	}
}

/**
 * Project models.
 */

type ProjectModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	CreationDate   types.String `tfsdk:"creation_date"`
	RevisionDate   types.String `tfsdk:"revision_date"`
}

type JSONProject struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	Name           string `json:"name"`
	CreationDate   string `json:"creationDate"`
	RevisionDate   string `json:"revisionDate"`
}

func (j *JSONProject) Parse() ProjectModel {
	return ProjectModel{
		ID:             types.StringValue(j.ID),
		OrganizationID: types.StringValue(j.OrganizationID),
		Name:           types.StringValue(j.Name),
		CreationDate:   types.StringValue(j.CreationDate),
		RevisionDate:   types.StringValue(j.RevisionDate),
	}
}
