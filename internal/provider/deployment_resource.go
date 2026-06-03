package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DeploymentResource{}

// NewDeploymentResource deploys a Docker container/app onto one of your VMs over
// SSH — the Terraform equivalent of `vxcli deploy container`.
func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	client *Client
}

type DeploymentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Image         types.String `tfsdk:"image"`
	Host          types.String `tfsdk:"host"`
	SSHUser       types.String `tfsdk:"ssh_user"`
	KeyPairName   types.String `tfsdk:"key_pair_name"`
	Ports         types.List   `tfsdk:"ports"`
	Env           types.List   `tfsdk:"env"`
	RestartPolicy types.String `tfsdk:"restart_policy"`
	Network       types.String `tfsdk:"network"`
	Command       types.String `tfsdk:"command"`
	EnableSSL     types.Bool   `tfsdk:"enable_ssl"`
	Domain        types.String `tfsdk:"domain"`
	SSLEmail      types.String `tfsdk:"ssl_email"`
	SessionID     types.String `tfsdk:"session_id"`
}

func (r *DeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Deploys a Docker container/app onto a VM over SSH (equivalent to `vxcli deploy container`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Deployment identifier (the container name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "Container name on the target host.",
				PlanModifiers: replace,
			},
			"image": schema.StringAttribute{
				Required:    true,
				Description: "Docker image to run (e.g. redis:7, nginx:latest, ghcr.io/acme/api:1.2).",
			},
			"host": schema.StringAttribute{
				Required:      true,
				Description:   "Target VM IP or hostname to deploy onto.",
				PlanModifiers: replace,
			},
			"ssh_user": schema.StringAttribute{
				Required:    true,
				Description: "SSH user on the target VM (e.g. ubuntu, root).",
			},
			"key_pair_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the SSH key-pair stored in your vxcloud Vault (server-side credential).",
			},
			"ports": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Port mappings, host:container (e.g. [\"80:8000\", \"6379:6379\"]).",
			},
			"env": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Environment variables as KEY=VALUE strings.",
			},
			"restart_policy": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("unless-stopped"),
				Description: "Docker restart policy. Defaults to unless-stopped.",
			},
			"network": schema.StringAttribute{
				Optional:    true,
				Description: "Docker network to attach the container to.",
			},
			"command": schema.StringAttribute{
				Optional:    true,
				Description: "Override the container's default command.",
			},
			"enable_ssl": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Provision a Let's Encrypt certificate and reverse-proxy for `domain`.",
			},
			"domain": schema.StringAttribute{
				Optional:    true,
				Description: "Public domain to route to the container (required when enable_ssl = true).",
			},
			"ssl_email": schema.StringAttribute{
				Optional:    true,
				Description: "Email for Let's Encrypt registration (required when enable_ssl = true).",
			},
			"session_id": schema.StringAttribute{
				Computed:    true,
				Description: "Deploy session id returned by the platform (for log streaming).",
			},
		},
	}
}

func (r *DeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("expected *Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

// listToCSV flattens a types.List of strings into a comma-joined value, the
// shape the node's multipart endpoint expects for ports/environment_vars.
func listToCSV(ctx context.Context, l types.List) (string, error) {
	if l.IsNull() || l.IsUnknown() {
		return "", nil
	}
	var items []string
	if diags := l.ElementsAs(ctx, &items, false); diags.HasError() {
		return "", fmt.Errorf("could not read list attribute")
	}
	return strings.Join(items, ","), nil
}

func (r *DeploymentResource) deployFields(ctx context.Context, plan DeploymentResourceModel) (map[string]string, error) {
	ports, err := listToCSV(ctx, plan.Ports)
	if err != nil {
		return nil, err
	}
	env, err := listToCSV(ctx, plan.Env)
	if err != nil {
		return nil, err
	}
	restart := plan.RestartPolicy.ValueString()
	if restart == "" {
		restart = "unless-stopped"
	}
	fields := map[string]string{
		"image":            plan.Image.ValueString(),
		"container_name":   plan.Name.ValueString(),
		"hostname":         plan.Host.ValueString(),
		"ssh_username":     plan.SSHUser.ValueString(),
		"key_pair_name":    plan.KeyPairName.ValueString(),
		"ports":            ports,
		"environment_vars": env,
		"restart_policy":   restart,
		"network":          plan.Network.ValueString(),
		"command":          plan.Command.ValueString(),
		"cloud_provider":   "docker",
	}
	if plan.EnableSSL.ValueBool() {
		fields["enable_ssl"] = "true"
		fields["domain"] = plan.Domain.ValueString()
		fields["ssl_email"] = plan.SSLEmail.ValueString()
	}
	return fields, nil
}

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.EnableSSL.ValueBool() && (plan.Domain.ValueString() == "" || plan.SSLEmail.ValueString() == "") {
		resp.Diagnostics.AddError("SSL misconfigured", "enable_ssl = true requires both domain and ssl_email to be set.")
		return
	}

	fields, err := r.deployFields(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid deployment input", err.Error())
		return
	}

	out, err := r.client.PostMultipart(ctx, "/api/v2/tenant/container/deploy", fields)
	if err != nil {
		resp.Diagnostics.AddError("Container deploy failed", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Name.ValueString())
	plan.SessionID = types.StringValue(firstString(out, "session_id", "sessionId", "id"))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Container deploys are imperative over SSH; the platform does not expose a
	// stable GET to reconcile arbitrary container state, so we keep the recorded
	// state as-is. (Drift on the host is surfaced on the next apply.)
	var state DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-deploy in place: the node pulls the (possibly new) image, recreates the
	// container with the updated env/ports, and keeps the same name/host.
	fields, err := r.deployFields(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid deployment input", err.Error())
		return
	}
	out, err := r.client.PostMultipart(ctx, "/api/v2/tenant/container/deploy", fields)
	if err != nil {
		resp.Diagnostics.AddError("Container re-deploy failed", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.Name.ValueString())
	plan.SessionID = types.StringValue(firstString(out, "session_id", "sessionId", "id"))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Best-effort teardown: ask the node to stop & remove the container on the
	// target host. A non-fatal warning is emitted if the host is unreachable so
	// the resource still leaves Terraform state.
	fields := map[string]string{
		"container_name": state.Name.ValueString(),
		"hostname":       state.Host.ValueString(),
		"ssh_username":   state.SSHUser.ValueString(),
		"key_pair_name":  state.KeyPairName.ValueString(),
		"cloud_provider": "docker",
	}
	if _, err := r.client.PostMultipart(ctx, "/api/v2/tenant/container/remove", fields); err != nil {
		resp.Diagnostics.AddWarning(
			"Container remove did not complete",
			fmt.Sprintf("Removed from Terraform state, but the platform reported: %s", err.Error()),
		)
	}
}
