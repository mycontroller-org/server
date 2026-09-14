import objectPath from "object-path"
import { DataType, FieldType } from "../../../../Constants/Form"

// get PhilipsHue Items
export const getPhilipsHueItems = (rootObject) => {
  objectPath.set(rootObject, "provider.syncInterval", "15m", true)
  const items = [
    {
      label: "configuration",
      fieldId: "!provider.configuration",
      fieldType: FieldType.Divider,
    },
    {
      label: "bridge",
      fieldId: "provider.host",
      isRequired: true,
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      validator: {
        isURL: {
          protocols: ["http", "https"],
        },
      },
      helperTextInvalid: "helper_text.invalid_url",
    },
    {
      label: "username",
      fieldId: "provider.username",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
      isRequired: true,
      helperText: "",
      helperTextInvalid: "helper_text.invalid_username",
      validated: "default",
      validator: { isLength: { min: 1, max: 100 }, isNotEmpty: {} },
    },
    {
      label: "sync_interval",
      fieldId: "provider.syncInterval",
      fieldType: FieldType.Text,
      dataType: DataType.String,
      value: "",
    },
  ]
  return items
}
