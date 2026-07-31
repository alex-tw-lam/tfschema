variable "nested" {
  default = {
    outer = {
      inner = "value"
      count = 2
    }
  }
}
