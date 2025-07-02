variable "mixed_payload" {
  description = "A complex tuple with nested objects and indexed validation."
  type = tuple([
    string,  // An ID
    number,  // A count
    object({ // A configuration block
      name    = string
      enabled = bool
      retries = number
    })
  ])

  validation {
    # Validate the first element (string)
    condition     = length(var.mixed_payload[0]) == 36
    error_message = "The first element must be a 36-character UUID."
  }

  validation {
    # Validate the second element (number)
    condition     = var.mixed_payload[1] >= 1 && var.mixed_payload[1] <= 8
    error_message = "The second element must be a count between 1 and 8."
  }

  validation {
    # Validate a nested attribute in the third element (object)
    condition     = var.mixed_payload[2].retries >= 0 && var.mixed_payload[2].retries <= 3
    error_message = "Retries in the third element's config must be between 0 and 3."
  }
} 
