variable "map_length_advanced_var" {
  type = map(object({
    id = string
  }))
  validation {
    condition     = length(var.map_length_advanced_var) > 0
    error_message = "Map must not be empty."
  }
}
