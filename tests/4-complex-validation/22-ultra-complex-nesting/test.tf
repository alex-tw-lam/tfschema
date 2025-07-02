variable "ultra_complex_structure" {
  description = "A highly complex and deeply nested structure designed to test the limits of the converter."
  type = object({
    environments = list(object({
      name          = string
      feature_flags = map(bool)
      deployment_config = tuple([
        string,      // region
        number,      // replica count
        set(string), // availability zones
        object({
          storage_type = string
          storage_size = number
        })
      ])
    })),
    service_endpoints = object({
      api  = tuple([string, number]),
      docs = tuple([string, number])
    }),
    auditors = optional(tuple([
      string, // auditor group name
      set(object({
        username = string
        level    = number
      }))
    ]))
  })

  # --- VALIDATION RULES ---
  validation {
    condition     = length(var.ultra_complex_structure.environments) > 0 # minItems
    error_message = "At least one environment must be configured."
  }

  validation {
    condition     = contains(["production", "staging"], var.ultra_complex_structure.environments[0].name) # enum
    error_message = "The first environment's name must be 'production' or 'staging'."
  }

  validation {
    condition     = var.ultra_complex_structure.environments[0].deployment_config[1] > 0 # minimum
    error_message = "Replica count for the first environment must be positive."
  }

  validation {
    condition     = var.ultra_complex_structure.environments[0].deployment_config[3].storage_size > 100 # minimum
    error_message = "Storage size for the first environment must be greater than 100."
  }

  validation {
    condition     = can(regex("^https", var.ultra_complex_structure.service_endpoints.docs[0])) # pattern
    error_message = "Docs endpoint URL must start with https."
  }

  validation {
    condition     = length(var.ultra_complex_structure.auditors[0]) > 3 # minLength
    error_message = "Auditor group name must be longer than 3 characters."
  }
} 
