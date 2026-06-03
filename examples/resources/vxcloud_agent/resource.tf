# Create an AgentControl agent.
# This is the Terraform equivalent of `vxcli agentcontrol agent create`.
#
# Requires `tenant_id` (X-Tenant-ID) — set it on the provider block or override
# per-resource below.

resource "vxcloud_agent" "compliance_copilot" {
  name        = "compliance-copilot"
  agent_type  = "rag"
  model       = "compliancellm"
  description = "FinTech compliance Q&A over policy documents."

  system_prompt = <<-EOT
    You are a FinTech compliance assistant. Answer only from the retrieved
    policy context. Cite the source section for every claim, and say "not in
    policy" when the context does not cover the question.
  EOT

  # Optional: override the provider-level tenant for this agent.
  # tenant_id = "00000000-0000-0000-0000-000000000000"
}

output "agent_id" {
  value = vxcloud_agent.compliance_copilot.id
}
