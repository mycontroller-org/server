// Operator values
export const Operator = {
  Equal: "eq",
  NotEqual: "ne",
  In: "in",
  NotIn: "nin",
  RangeIn: "range_in",
  RangeNotIn: "range_not_in",
  GreaterThan: "gt",
  LessThan: "lt",
  GreaterThanEqual: "gte",
  LessThanEqual: "lte",
  Exists: "exists",
  Regex: "regex",
}

// Operator options list
export const OperatorOptions = [
  { value: Operator.Equal, label: "==", description: "Equal" },
  { value: Operator.NotEqual, label: "!=", description: "Not equal" },
  { value: Operator.GreaterThan, label: ">", description: "greater than" },
  { value: Operator.GreaterThanEqual, label: ">=", description: "greater than equal" },
  { value: Operator.LessThan, label: "<", description: "Less than" },
  { value: Operator.LessThanEqual, label: "<=", description: "Less than equal" },
  { value: Operator.In, label: "In", description: "in" },
  { value: Operator.NotIn, label: "Not In", description: "not in" },
  { value: Operator.RangeIn, label: "In Range", description: "in range" },
  { value: Operator.RangeNotIn, label: "Not In Range", description: "not in range" },
  { value: Operator.Regex, label: "Regex", description: "Regex" },
  { value: Operator.Exists, label: "Exists", description: "Exists" },
]
