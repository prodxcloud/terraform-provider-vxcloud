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
	TenantID types.String `tfsdk:"tenant_id"`
	Username types.String `tfsdk:"username"`
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
				Description: "Developer API key (xc_dev_/xc_live_) sent as X-API-Key. May also be set via VXCLOUD_API_TOKEN or VXCLOUD_API_KEY.",
				Optional:    true,
				Sensitive:   true,
			},
			"endpoint": schema.StringAttribute{
				Description: "Tenant node base URL where deploys and agentcontrol run. Defaults to https://node1.vxcloud.io. May also be set via VXCLOUD_ENDPOINT.",
				Optional:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "Tenant id (X-Tenant-ID) for agentcontrol resources. May also be set via VXCLOUD_TENANT_ID.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username (X-Username) for agentcontrol resources. Defaults to email. May also be set via VXCLOUD_USERNAME.",
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
		apiToken = firstEnv("VXCLOUD_API_TOKEN", "VXCLOUD_API_KEY")
	}
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("VXCLOUD_ENDPOINT")
	}
	if endpoint == "" {
		endpoint = "https://node1.vxcloud.io"
	}
	tenantID := data.TenantID.ValueString()
	if tenantID == "" {
		tenantID = os.Getenv("VXCLOUD_TENANT_ID")
	}
	username := data.Username.ValueString()
	if username == "" {
		username = os.Getenv("VXCLOUD_USERNAME")
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing api_token",
			"Set the api_token attribute or the VXCLOUD_API_TOKEN / VXCLOUD_API_KEY environment variable.",
		)
		return
	}

	client := NewClient(endpoint, email, apiToken)
	client.TenantID = tenantID
	client.Username = username
	resp.DataSourceData = client
	resp.ResourceData = client
}

// firstEnv returns the first non-empty environment variable among keys.
func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func (p *VxCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDeploymentResource,
		NewAgentResource,
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
