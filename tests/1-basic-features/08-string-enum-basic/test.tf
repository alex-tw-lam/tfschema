variable "string_enum_var" {
  type    = string
  validation {
    condition     = contains(["cat", "dog", "fish"], var.string_enum_var)
    error_message = "Allowed values are cat, dog, or fish."
  }
}
