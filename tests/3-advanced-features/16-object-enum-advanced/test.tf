variable "object_enum_advanced_var" {
  type = object({
    config = object({
      setting = string
    })
  })
  validation {
    condition     = contains(["mode1", "mode2"], var.object_enum_advanced_var.config.setting)
    error_message = "Setting must be 'mode1' or 'mode2'."
  }
}
