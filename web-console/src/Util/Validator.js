import v from "validator"
import { t } from "typy"
import { isDate } from "moment"
import { getValue } from "./Util"
import { DataType, FieldType } from "../Constants/Form"

const isLengthDualList = (val, opts) => {
  if (t(opts, "min").isNumber && val.right.length < opts.min) {
    return false
  }
  if (t(opts, "max").isNumber && val.right.length > opts.max) {
    return false
  }
  return true
}

const isLengthArray = (val, opts) => {
  if (t(val).isArray) {
    if (t(opts, "min").isNumber && val.length < opts.min) {
      return false
    }
    if (t(opts, "max").isNumber && val.length > opts.max) {
      return false
    }
    return true
  }
  return false
}

const baudRates = [
  50, 75, 110, 134, 150, 200, 300, 600, 1200, 1800, 2400, 4800, 9600, 19200, 38400, 57600, 115200, 230400,
  460800, 500000, 576000, 921600, 1000000, 1152000, 1500000, 2000000, 2500000, 3000000, 3500000, 4000000,
]

const isBaudRate = (val) => {
  const value = v.toInt(val)
  return baudRates.includes(value)
}

const isLabelKey = (val) => {
  const regexIsLabelKey = /^[a-z0-9_/-]+$/
  return regexIsLabelKey.test(val)
}

export const validate = (func, val, opts) => {
  switch (func) {
    case "isEmail":
      return v.isEmail(val, opts)

    case "isInt":
    case "isInteger":
      return v.isInt(val, opts)

    case "isDecimal":
      return v.isDecimal(val, opts)

    case "isFloat":
      return v.isFloat(val, opts)

    case "isCurrency":
      if (parseFloat(val) === 0.0) {
        // workaround for currency should not be zero
        return false
      }
      return v.isCurrency(val, opts)

    case "isEmpty":
      // validator.isEmpty only accepts strings; arrays/objects must be handled here
      if (val === undefined || val === null) {
        return true
      }
      if (t(val).isArray) {
        return val.length === 0
      }
      if (t(val).isObject) {
        return Object.keys(val).length === 0
      }
      return v.isEmpty(String(val), opts)

    case "isNotEmpty":
      if (val === undefined || val === null) {
        return false
      }
      if (t(val).isArray) {
        return val.length > 0
      }
      if (t(val).isObject) {
        return Object.keys(val).length > 0
      }
      return !v.isEmpty(String(val), opts)

    case "isEqual":
      return val === opts.value

    case "isAlphanumeric":
      return v.isAlphanumeric(val)

    case "isLength":
      return v.isLength(val, opts)

    case "isLengthDualList":
      return isLengthDualList(val, opts)

    case "isObject":
      return t(val).isObject

    case "isDate":
      return isDate(val)

    case "isLengthArray":
      return isLengthArray(val, opts)

    case "isURL":
      return v.isURL(val, opts)

    case "isBaudRate":
      return isBaudRate(val)

    case "isLabel":
      if (val === undefined || val === null) {
        return false
      }
      const keys = Object.keys(val)
      for (let index = 0; index < keys.length; index++) {
        const key = keys[index]
        const isValid = isLabelKey(key)
        if (!isValid) {
          return false
        }
        // if (val[key] === ""){
        //   return false
        // }
      }
      return true

    case "isID":
      const regexIsID = /^[a-zA-Z0-9_-]+$/
      return regexIsID.test(val)

    case "isLabelKey":
      return isLabelKey(val)

    case "isKey":
      const regexIsKey = /^[a-zA-Z0-9\._\/-]+$/
      return regexIsKey.test(val)

    case "isVariableKey":
      const regexIsVariableKey = /^[a-zA-Z0-9_]+$/
      return regexIsVariableKey.test(val)

    case "isVersion":
      const regexIsVersion = /^[a-z0-9\._-]+$/
      return regexIsVersion.test(val)

    case "isPort":
      return v.isPort(val)
    default:
      return true
  }
}

export const validateItem = (item, value) => {
  const isRequired = getValue(item, "isRequired", false)
  const fieldType = getValue(item, "fieldType", "")

  if (isRequired) {
    if (fieldType !== FieldType.Switch && validate("isEmpty", value, {})) {
      return false
    }
  } else if (fieldType !== FieldType.Switch && validate("isEmpty", value, {})) {
    // optional empty values skip format checks (e.g. optional email)
    return true
  }

  // Array fields: never run string-only validators (validator lib throws on non-strings)
  switch (fieldType) {
    case FieldType.KeyValueMap:
    case FieldType.Labels:
    case FieldType.VariablesMap:
    case FieldType.ChartYAxisConfigMap:
      const validateKeyFnKV = getValue(item, "validateKeyFunc", null)
      const validateValueFnKV = getValue(item, "validateValueFunc", null)
      if (validateKeyFnKV !== null || validateValueFnKV !== null) {
        const valuesKV = value
        const keysKV = Object.keys(valuesKV)
        for (let index = 0; index < keysKV.length; index++) {
          const key = keysKV[index]
          const value = valuesKV[key]

          if (validateKeyFnKV !== null && !validateKeyFnKV(key)) {
            return false
          }
          if (validateValueFnKV !== null && !validateValueFnKV(value)) {
            return false
          }
        }
      }

      break

    case FieldType.DynamicArray:
      {
        if (!Array.isArray(value)) {
          return !isRequired
        }
        if (isRequired && value.length === 0) {
          return false
        }
        const validateValueFnDA = getValue(item, "validateValueFunc", null)
        if (validateValueFnDA !== null) {
          for (let index = 0; index < value.length; index++) {
            if (!validateValueFnDA(value[index])) {
              return false
            }
          }
        }
      }
      break

    case FieldType.SelectTypeAhead:
    case FieldType.SelectTypeAheadMultiple:
    case FieldType.SelectTypeAheadAsync:
      if (item.dataType === DataType.ArrayString || Array.isArray(value)) {
        if (!Array.isArray(value)) {
          return !isRequired
        }
        if (isRequired && value.length === 0) {
          return false
        }
        if (item.validator) {
          const funcs = Object.keys(item.validator)
          for (let index = 0; index < funcs.length; index++) {
            const func = funcs[index]
            if (!validate(func, value, item.validator[func])) {
              return false
            }
          }
        }
        return true
      }
      // single string select
      if (item.validator) {
        const funcs = Object.keys(item.validator)
        for (let index = 0; index < funcs.length; index++) {
          const func = funcs[index]
          if (!validate(func, value, item.validator[func])) {
            return false
          }
        }
      }
      break

    case FieldType.PolicyStatements:
      if (!Array.isArray(value) || value.length === 0) {
        return !isRequired
      }
      for (let index = 0; index < value.length; index++) {
        const st = value[index] || {}
        if (!Array.isArray(st.actions) || st.actions.length === 0) {
          return false
        }
        if (!Array.isArray(st.resources) || st.resources.length === 0) {
          return false
        }
      }
      break

    default:
      if (!item.validator) {
        return true
      }
      const funcs = Object.keys(item.validator)
      for (let index = 0; index < funcs.length; index++) {
        const func = funcs[index]
        if (!validate(func, value, item.validator[func])) {
          return false
        }
      }
      break
  }

  return true
}
