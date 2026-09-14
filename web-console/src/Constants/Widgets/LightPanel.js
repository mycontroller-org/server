// light panel types
export const LightType = {
  Single: "light_single",
  CWWW: "light_cw_ww",
  RGB: "light_rgb",
  RGBCW: "light_rgbcw",
  RGBCWWW: "light_rgbcwww",
}

// light panel options
export const LightTypeOptions = [
  {
    value: LightType.Single,
    label: "opts.light_type.label_single",
    description: "opts.light_type.desc_single",
  },
  { value: LightType.CWWW, label: "opts.light_type.label_cw_ww", description: "opts.light_type.desc_cw_ww" },
  { value: LightType.RGB, label: "opts.light_type.label_rgb", description: "opts.light_type.desc_rgb" },
  {
    value: LightType.RGBCW,
    label: "opts.light_type.label_rgb_cw",
    description: "opts.light_type.desc_rgb_cw",
  },
  {
    value: LightType.RGBCWWW,
    label: "opts.light_type.label_rgb_cw_ww",
    description: "opts.light_type.desc_rgb_cw_ww",
  },
]

// RGB components
export const RGBComponentType = {
  ColorPickerQuick: "rgb_color_picker",
  HueSlider: "hue_slider",
}

// RGB component options
export const RGBComponentOptions = [
  {
    value: RGBComponentType.ColorPickerQuick,
    label: "opts.rgb_component.label_color_picker",
    description: "opts.rgb_component.desc_color_picker",
  },
  {
    value: RGBComponentType.HueSlider,
    label: "opts.rgb_component.label_hue_slider",
    description: "opts.rgb_component.desc_hue_slider",
  },
]
