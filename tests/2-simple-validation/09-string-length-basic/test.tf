variable "string_length_var" {
  type    = string
  validation {
    condition     = length(var.string_length_var) >= 3 && length(var.string_length_var) <= 10
    error_message = "String length must be between 3 and 10."
  }
}
