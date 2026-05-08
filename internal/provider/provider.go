package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &VxCloudProvider{}

type VxCloudProvider struct {
	version string
}

type VxCloudProviderModel struct {
	Email    types.String `tfsdk:"email"`
	APIToken types.String `tfsdk:"api_token"`
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *VxCloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vxcloud"
	resp.Version = p.version
}

func (p *VxCloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage resources on the vxcloud platform.",
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Description: "Email associated with your vxcloud account. May also be set via VXCLOUD_EMAIL.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "API token for the vxcloud API. May also be set via VXCLOUD_API_TOKEN.",
				Optional:    true,
				Sensitive:   true,
			},
			"endpoint": schema.StringAttribute{
				Description: "Override the vxcloud API endpoint. Defaults to https://api.vxcloud.com.",
				Optional:    true,
			},
		},
	}
}

func (p *VxCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data VxCloudProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := data.Email.ValueString()
	if email == "" {
		email = os.Getenv("VXCLOUD_EMAIL")
	}
	apiToken := data.APIToken.ValueString()
	if apiToken == "" {
		apiToken = os.Getenv("VXCLOUD_API_TOKEN")
	}
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = "https://api.vxcloud.com"
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing api_token",
			"Set the api_token attribute or the VXCLOUD_API_TOKEN environment variable.",
		)
		return
	}

	client := NewClient(endpoint, email, apiToken)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *VxCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRedisResource,
	}
}

func (p *VxCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VxCloudProvider{version: version}
	}
}
