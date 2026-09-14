// image panel types
export const ImageSourceType = {
  Field: "image_from_field",
  URL: "url",
  Disk: "disk",
}

// image panel options
export const ImageSourceTypeOptions = [
  {
    value: ImageSourceType.Field,
    label: "opts.image_source.label_field",
    description: "opts.image_source.desc_field",
  },
  {
    value: ImageSourceType.URL,
    label: "opts.image_source.label_url",
    description: "opts.image_source.desc_url",
  },
  {
    value: ImageSourceType.Disk,
    label: "opts.image_source.label_disk",
    description: "opts.image_source.desc_disk",
  },
]

// image types in field
export const ImageType = {
  Image: "image",
  CustomMapping: "custom_mapping",
}

// image type options
export const ImageTypeOptions = [
  {
    value: ImageType.Image,
    label: "opts.image_load_type.label_image",
    description: "opts.image_load_type.desc_image",
  },
  {
    value: ImageType.CustomMapping,
    label: "opts.image_load_type.label_custom_mapping",
    description: "opts.image_load_type.desc_custom_mapping",
  },
]

export const ICON_PREFIX = "icon:"

// icon types
export const IconType = {
  // https://github.com/react-icons/react-icons
  AntDesignIcons: "ai",
  BootstrapIcons: "bs",
  BoxIcons: "bi",
  Devicons: "di",
  Feather: "fi",
  FlatColorIcons: "fc",
  FontAwesomeIcons: "fa",
  GameIcons: "gi",
  GithubOcticonsIcons: "go",
  GrommetIcons: "gr",
  Heroicons: "hi",
  IcoMoonFree: "im",
  Ionicons4: "io",
  Ionicons5: "io5",
  MaterialDesignIcons: "md",
  RemixIcon: "ri",
  SimpleIcons: "si",
  Typicons: "ti",
  VSCodeIcons: "vsc",
  WeatherIcons: "ws",
  CSS_gg: "cg",
  // https://github.com/patternfly/patternfly-react/tree/main/packages/react-icons
  PatternFlyIcons: "pf",
}

// image rotation types
export const ImageRotationType = {
  Rotate_0: "0",
  RotateRight_90: "90",
  RotateLeft_90: "-90",
  Rotate_180: "180",
}

// image rotation options
export const ImageRotationTypeOptions = [
  {
    value: ImageRotationType.Rotate_0,
    label: "opts.image_rotation.label_rotate_none",
    description: "opts.image_rotation.desc_rotate_none",
  },
  {
    value: ImageRotationType.RotateRight_90,
    label: "opts.image_rotation.label_rotate_right_90",
    description: "opts.image_rotation.desc_rotate_right_90",
  },
  {
    value: ImageRotationType.RotateLeft_90,
    label: "opts.image_rotation.label_rotate_left_90",
    description: "opts.image_rotation.desc_rotate_left_90",
  },
  {
    value: ImageRotationType.Rotate_180,
    label: "opts.image_rotation.label_rotate_180",
    description: "opts.image_rotation.desc_rotate_180",
  },
]
