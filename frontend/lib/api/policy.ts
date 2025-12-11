import { authClient } from "@/lib/api/client";
import type { PolicyLimits } from "@/lib/components/schemas";

// Public endpoint to fetch limited system policy (no admin token needed).
export function getPolicyLimits(): Promise<PolicyLimits> {
  return authClient.get<PolicyLimits>("/api/policy/limits");
}

