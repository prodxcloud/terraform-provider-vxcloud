package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RedisResource{}

func NewRedisResource() resource.Resource {
	return &RedisResource{}
}

type RedisResource struct {
	client *Client
}

type RedisResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	ServerName   types.String `tfsdk:"server_name"`
	ServerType   types.String `tfsdk:"server_type"`
	Datacenter   types.String `tfsdk:"datacenter"`
	SupportLevel types.String `tfsdk:"support_level"`
}

func (r *RedisResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis"
}

func (r *RedisResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Redis service on vxcloud.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "Project the Redis service belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_name": schema.StringAttribute{
				Required:    true,
				Description: "Display name for the Redis server.",
			},
			"server_type": schema.StringAttribute{
				Required:    true,
				Description: "Server flavor (e.g. SMALL-2C, MEDIUM-4C).",
			},
			"datacenter": schema.StringAttribute{
				Required:    true,
				Description: "Datacenter region (e.g. fsn).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"support_level": schema.StringAttribute{
				Required:    true,
				Description: "Support tier (e.g. level1, level2).",
			},
		},
	}
}

func (r *RedisResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("expected *Client, got %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *RedisResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RedisResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: call vxcloud API to create the Redis service.
	plan.ID = types.StringValue("redis-stub-id")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RedisResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RedisResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO: call vxcloud API to refresh the resource.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RedisResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RedisResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO: call vxcloud API to update.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RedisResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RedisResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO: call vxcloud API to delete.
}
