variable "bool_enum_var" {
  type    = bool
  validation {
    condition     = contains([true], var.bool_enum_var)
    error_message = "Value must be true."
  }
}
