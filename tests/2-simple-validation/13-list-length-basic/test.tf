variable "list_length_var" {
  type    = list(string)
  validation {
    condition     = length(var.list_length_var) == 3
    error_message = "List must contain exactly 3 elements."
  }
}
