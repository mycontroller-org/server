// control panel types
export const ControlType = {
  SwitchToggle: "switch_toggle",
  SwitchButton: "switch_button",
  MixedControl: "mixed_control",
}

// utilization panel type options
export const ControlTypeOptions = [
  { value: ControlType.SwitchToggle, label: "opts.control_type.label_switch_toggle" },
  { value: ControlType.SwitchButton, label: "opts.control_type.label_switch_button" },
  { value: ControlType.MixedControl, label: "opts.control_type.label_mixed_control" },
]

// switch panel button types
export const ButtonType = {
  Primary: "primary",
  Secondary: "secondary",
  Warning: "warning",
  Danger: "danger",
}

// switch panel button type options
export const ButtonTypeOptions = [
  { value: ButtonType.Primary, label: "opts.button_type.label_primary" },
  { value: ButtonType.Secondary, label: "opts.button_type.label_secondary" },
  { value: ButtonType.Warning, label: "opts.button_type.label_warning" },
  { value: ButtonType.Danger, label: "opts.button_type.label_danger" },
]

// mixed control types
export const MixedControlType = {
  ToggleSwitch: "switch_toggle",
  ButtonSwitch: "switch_button",
  PushButton: "button_push",
  SelectOptions: "options_select",
  TabOptions: "options_tab",
  Input: "input",
  Slider: "slider",
}

// mixed control type options
export const MixedControlTypeOptions = [
  { value: MixedControlType.ToggleSwitch, label: "opts.control_type_mixed.label_switch_toggle" },
  { value: MixedControlType.ButtonSwitch, label: "opts.control_type_mixed.label_switch_button" },
  { value: MixedControlType.PushButton, label: "opts.control_type_mixed.label_button_push" },
  { value: MixedControlType.SelectOptions, label: "opts.control_type_mixed.label_options_select" },
  { value: MixedControlType.TabOptions, label: "opts.control_type_mixed.label_options_tab" },
  { value: MixedControlType.Input, label: "opts.control_type_mixed.label_input" },
  { value: MixedControlType.Slider, label: "opts.control_type_mixed.label_slider" },
]
