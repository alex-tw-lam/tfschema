variable "map_enum_advanced_var" {
  type = map(string)
  validation {
    condition = alltrue([
      for k, v in var.map_enum_advanced_var : contains(["a", "b", "c"], v)
    ])
    error_message = "All values must be a, b, or c."
  }
}
