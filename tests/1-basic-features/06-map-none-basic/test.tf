variable "map_var" {
  type        = map(string)
  description = "A basic map variable."
  default = {
    key1 = "value1"
    key2 = "value2"
  }
}
