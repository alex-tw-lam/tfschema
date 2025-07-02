variable "number_enum_var" {
  type    = number
  validation {
    condition     = contains([1, 2, 3], var.number_enum_var)
    error_message = "Allowed values are 1, 2, or 3."
  }
}
