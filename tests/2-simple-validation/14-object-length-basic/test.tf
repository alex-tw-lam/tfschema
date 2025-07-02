variable "object_length_var" {
  type    = object({
    a = string
    b = string
    c = optional(string)
  })
  validation {
    # This isn't directly translatable to JSON schema object property counts
    # but represents a constraint on the variable itself.
    # The converter should ideally handle this gracefully.
    condition     = length(keys(var.object_length_var)) >= 2
    error_message = "Object must have at least 2 properties."
  }
}
