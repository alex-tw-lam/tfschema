variable "list_enum_advanced_var" {
  type = list(object({
    name  = string
    value = string
  }))
  validation {
    condition = alltrue([
      for item in var.list_enum_advanced_var : contains(["allow1", "allow2"], item.value)
    ])
    error_message = "All item values must be 'allow1' or 'allow2'."
  }
}
