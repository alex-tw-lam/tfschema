variable "number_range_var" {
  type    = number
  validation {
    condition     = var.number_range_var > 0 && var.number_range_var <= 100
    error_message = "Number must be between 1 and 100."
  }
}
