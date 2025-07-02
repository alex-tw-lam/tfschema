# Terraschema custom-validation test case
# Based on terraschema/test/modules/custom-validation/variables.tf

variable "a_string_enum_kind_1" {
  type        = string
  default     = "a"
  description = "A string variable that must be one of the values 'a', 'b', or 'c'"
  validation {
    condition     = contains(["a", "b", "c"], var.a_string_enum_kind_1)
    error_message = "Invalid value for a_string_enum_kind_1"
  }
}

variable "a_string_enum_kind_2" {
  type        = string
  default     = "a"
  description = "A string variable that must be one of the values 'a', 'b', or 'c'"
  validation {
    condition     = var.a_string_enum_kind_2 == "a" || var.a_string_enum_kind_2 == "b" || var.a_string_enum_kind_2 == "c"
    error_message = "Invalid value for a_string_enum_kind_2"
  }
}

variable "a_number_enum_kind_1" {
  type        = number
  default     = 1
  description = "A number variable that must be one of the values 1, 2, or 3"
  validation {
    condition     = contains([1, 2, 3], var.a_number_enum_kind_1)
    error_message = "Invalid value for a_number_enum_kind_1"
  }
}

variable "a_number_enum_kind_2" {
  type        = number
  default     = 1
  description = "A number variable that must be one of the values 1, 2, or 3"
  validation {
    condition     = var.a_number_enum_kind_2 == 1 || var.a_number_enum_kind_2 == 2 || var.a_number_enum_kind_2 == 3
    error_message = "Invalid value for a_number_enum_kind_2"
  }
}

variable "a_number_exclusive_maximum_minimum" {
  type        = number
  default     = 1
  description = "A number variable that must be greater than 0 and less than 10"
  validation {
    condition     = var.a_number_exclusive_maximum_minimum > 0 && var.a_number_exclusive_maximum_minimum < 10
    error_message = "a_number_exclusive_maximum_minimum must be less than 10 and greater than 0"
  }
}

variable "a_number_maximum_minimum" {
  type        = number
  default     = 0
  description = "A number variable that must be between 0 and 10 (inclusive)"
  validation {
    condition     = var.a_number_maximum_minimum >= 0 && var.a_number_maximum_minimum <= 10
    error_message = "a_number_maximum_minimum must be less than or equal to 10 and greater than or equal to 0"
  }
}

variable "a_list_maximum_minimum_length" {
  type        = list(string)
  default     = ["a"]
  description = "A list variable that must have a length greater than 0 and less than 10"
  validation {
    condition     = length(var.a_list_maximum_minimum_length) > 0 && length(var.a_list_maximum_minimum_length) < 10
    error_message = "a_list_maximum_minimum_length must have a length greater than 0 and less than 10"
  }
}

variable "a_string_maximum_minimum_length" {
  type        = string
  description = "A string variable that must have a length less than 10 and greater than 0"
  validation {
    condition     = 0 < length(var.a_string_maximum_minimum_length) && length(var.a_string_maximum_minimum_length) < 10
    error_message = "a_string_maximum_minimum_length must have a length less than 10 and greater than 0"
  }
  default = "a"
}

variable "a_string_set_length" {
  type        = string
  description = "A string variable that must have length 4"
  validation {
    condition     = 4 == length(var.a_string_set_length)
    error_message = "a_string_set_length must have length 4"
  }
  default = "abcd"
}

variable "a_string_pattern_1" {
  type        = string
  description = "A string variable that must be a valid IPv4 address"
  validation {
    condition     = can(regex("^[0-9]{1,3}(\\.[0-9]{1,3}){3}$", var.a_string_pattern_1))
    error_message = "a_string_pattern_1 must be an IPv4 address"
  }
  default = "1.1.1.1"
}

variable "a_string_pattern_2" {
  type        = string
  description = "string that must be a valid colour hex code in the form #RRGGBB"
  validation {
    condition     = can(regex("^#[0-9a-fA-F]{6}$", var.a_string_pattern_2))
    error_message = "a_string_pattern_2 must be a valid colour hex code in the form #RRGGBB"
  }
  default = "#000000"
} 
