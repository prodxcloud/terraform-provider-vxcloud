package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AgentResource{}

// NewAgentResource manages an AgentControl agent — the Terraform equivalent of
// `vxcli agentcontrol agent create`.
func NewAgentResource() resource.Resource {
	return &AgentResource{}
}

type AgentResource struct {
	client *Client
}

type AgentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	AgentType    types.String `tfsdk:"agent_type"`
	Model        types.String `tfsdk:"model"`
	Description  types.String `tfsdk:"description"`
	SystemPrompt types.String `tfsdk:"system_prompt"`
	TenantID     types.String `tfsdk:"tenant_id"`
}

const agentCollection = "/api/v2/agentcontrol/agents"

func (r *AgentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *AgentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an AgentControl agent (equivalent to `vxcli agentcontrol agent create`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "AgentControl agent id assigned by the platform.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable agent name.",
			},
			"agent_type": schema.StringAttribute{
				Optional:    true,
				Description: "Agent type/kind (e.g. assistant, rag, tool-calling). Platform default applies when omitted.",
			},
			"model": schema.StringAttribute{
				Optional:    true,
				Description: "Backing model id for the agent (e.g. a deployed model or vxthinkingllm SLM).",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Free-text description of what the agent does.",
			},
			"system_prompt": schema.StringAttribute{
				Optional:    true,
				Description: "System prompt that defines the agent's behavior.",
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Description: "Override the provider-level tenant id (X-Tenant-ID) for this agent.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *AgentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// scoped returns the client honoring an optional per-resource tenant override.
func (r *AgentResource) scoped(m AgentResourceModel) *Client {
	return r.client.withTenant(m.TenantID.ValueString())
}

func agentPayload(m AgentResourceModel) map[string]any {
	p := map[string]any{"name": m.Name.ValueString()}
	if v := m.AgentType.ValueString(); v != "" {
		p["agent_type"] = v
	}
	if v := m.Model.ValueString(); v != "" {
		p["model"] = v
	}
	if v := m.Description.ValueString(); v != "" {
		p["description"] = v
	}
	if v := m.SystemPrompt.ValueString(); v != "" {
		p["system_prompt"] = v
	}
	return p
}

func (r *AgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AgentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.scoped(plan).PostJSON(ctx, agentCollection, agentPayload(plan))
	if err != nil {
		resp.Diagnostics.AddError("Agent create failed", err.Error())
		return
	}

	id := firstString(out, "id", "agent_id", "agentId", "_id", "uuid")
	if id == "" {
		resp.Diagnostics.AddError("Agent create returned no id", fmt.Sprintf("response: %v", out))
		return
	}
	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AgentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.scoped(state).Get(ctx, agentCollection+"/"+state.ID.ValueString())
	if err != nil {
		if ae, ok := err.(*apiError); ok && ae.Status == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Agent read failed", err.Error())
		return
	}

	// Best-effort refresh: only overwrite fields the API actually returns so we
	// don't clobber config-only values with empty strings.
	if v := firstString(out, "name"); v != "" {
		state.Name = types.StringValue(v)
	}
	if v := firstString(out, "agent_type"); v != "" {
		state.AgentType = types.StringValue(v)
	}
	if v := firstString(out, "model"); v != "" {
		state.Model = types.StringValue(v)
	}
	if v := firstString(out, "description"); v != "" {
		state.Description = types.StringValue(v)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AgentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.scoped(plan).PutJSON(ctx, agentCollection+"/"+plan.ID.ValueString(), agentPayload(plan))
	if err != nil {
		resp.Diagnostics.AddError("Agent update failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AgentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.scoped(state).Delete(ctx, agentCollection+"/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Agent delete failed", err.Error())
	}
}
