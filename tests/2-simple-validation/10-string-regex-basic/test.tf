variable "string_regex_var" {
  type    = string
  validation {
    condition     = can(regex("^[a-zA-Z0-9]+$", var.string_regex_var))
    error_message = "String must be alphanumeric."
  }
}
