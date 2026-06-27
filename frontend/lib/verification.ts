import { ApiError } from "@/lib/api"

export const EMAIL_VERIFICATION_REQUIRED_MESSAGE = "Verify your email in Settings to perform this action."

export function isEmailNotVerifiedError(err: unknown): boolean {
  return err instanceof ApiError && err.code === "AUTH_EMAIL_NOT_VERIFIED"
}
