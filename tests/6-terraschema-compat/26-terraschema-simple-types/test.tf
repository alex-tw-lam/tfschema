# Terraschema simple-types test case
# Based on terraschema/test/modules/simple-types/variables.tf

variable "a_string" {
  type        = string
  default     = "a string"
  description = "This is a string"
}

variable "a_number" {
  type        = number
  description = "This is a number"
}

variable "a_bool" {
  type        = bool
  default     = false
  description = "This is a boolean"
}

variable "a_nullable_string" {
  type        = string
  nullable    = true
  description = "This is a nullable string"
}

variable "a_list" {
  type        = list(string)
  default     = ["a", "b", "c"]
  description = "This is a list of strings"
}

variable "a_map_of_strings" {
  type = map(string)
  default = {
    a = "a"
    b = "b"
    c = "c"
  }
  description = "This is a map of strings"
}

variable "an_object" {
  type = object({
    a = string
    b = number
    c = bool
  })
  default = {
    a = "a"
    b = 1
    c = true
  }
  description = "This is an object"
}

variable "a_tuple" {
  type        = tuple([string, number, bool])
  default     = ["a", 1, true]
  description = "This is a tuple"
}

variable "a_set" {
  type        = set(string)
  default     = ["a", "b", "c"]
  description = "This is a set of strings"
} 
