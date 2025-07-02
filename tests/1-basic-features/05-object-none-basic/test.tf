variable "object_var" {
  type = object({
    name = string
    age  = number
  })
  description = "A basic object variable."
  default = {
    name = "John"
    age  = 30
  }
}
