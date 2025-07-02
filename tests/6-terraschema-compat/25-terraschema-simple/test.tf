# Terraschema simple test case
# Based on terraschema/test/modules/simple/variables.tf

variable "name" {
  type        = string
  description = "Your name."
  default     = "world"
}

variable "age" {
  type        = number
  description = "Your age. Required."
}
