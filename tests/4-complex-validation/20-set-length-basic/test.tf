variable "unique_values" {
  description = "A set of unique, non-empty strings (4-7 items)."
  type        = set(string)

  validation {
    condition     = length(var.unique_values) >= 4 && length(var.unique_values) <= 7
    error_message = "You must provide between 4 and 7 unique strings."
  }
}
