variable "complex_config" {
  description = "A highly nested and complex configuration object."
  type = object({
    # Simple types
    service_name = string
    is_enabled   = bool
    api_version  = number

    # String with regex and length validation
    cluster_prefix = string

    # Number with range validation
    instance_count = number

    # List of primitives with length validation
    availability_zones = list(string)

    # Set of objects with validation on nested fields
    user_identities = set(object({
      username     = string
      email        = string
      access_level = number
    }))

    # Map of complex objects with validation on nested fields
    component_settings = map(object({
      enabled   = bool
      retries   = optional(number, 3)
      timeout   = number
      endpoints = list(string)
    }))

    # Nested object with optional fields and nested optionals
    security_profile = object({
      firewall_enabled = bool
      allowed_ips      = optional(list(string))
      ports = optional(object({
        http  = optional(number, 80)
        https = optional(number, 443)
      }))
    })

    # Enum validation
    environment = string
  })

  validation {
    condition     = length(var.complex_config.service_name) > 0 && length(var.complex_config.service_name) <= 20
    error_message = "Service name must be between 1 and 20 characters."
  }

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.complex_config.cluster_prefix))
    error_message = "Cluster prefix can only contain lowercase letters, numbers, and hyphens."
  }

  validation {
    condition     = var.complex_config.instance_count >= 1 && var.complex_config.instance_count <= 10
    error_message = "Instance count must be between 1 and 10."
  }

  validation {
    condition     = length(var.complex_config.availability_zones) >= 2
    error_message = "At least two availability zones are required."
  }

  validation {
    condition = alltrue([
      for user in var.complex_config.user_identities : can(regex("^[a-z0-9_]{3,16}$", user.username))
    ])
    error_message = "All usernames must be 3-16 characters long and contain only lowercase letters, numbers, and underscores."
  }

  validation {
    condition = alltrue([
      for user in var.complex_config.user_identities : (user.access_level >= 1 && user.access_level <= 5)
    ])
    error_message = "Access level for all users must be between 1 and 5."
  }

  validation {
    condition = alltrue([
      for _, setting in var.complex_config.component_settings : setting.timeout > 0
    ])
    error_message = "Component timeout must be a positive number."
  }

  validation {
    condition     = contains(["development", "staging", "production"], var.complex_config.environment)
    error_message = "Environment must be one of: development, staging, or production."
  }
}
