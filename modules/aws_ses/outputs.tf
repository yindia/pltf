output "identity_arn" {
  value = aws_ses_domain_identity.email.arn
}

output "sender_policy_arn" {
  value = aws_iam_policy.sender.arn
}
