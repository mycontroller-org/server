import { QuestionIcon } from "@patternfly/react-icons"
import React from "react"
// import * as AntDesignIcons from "react-icons/ai"
// import * as BootstrapIcons from "react-icons/bs"
// import * as BoxIcons from "react-icons/bi"
// import * as Devicons from "react-icons/di"
// import * as Feather from "react-icons/fi"
// import * as FlatColorIcons from "react-icons/fc"
// import * as FontAwesome from "react-icons/fa"
// import * as GameIcons from "react-icons/gi"
// import * as GithubOcticonsIcons from "react-icons/go"
// import * as GrommetIcons from "react-icons/gr"
// import * as Heroicons from "react-icons/hi"
// import * as IcoMoonFree from "react-icons/im"
// import * as Ionicons4 from "react-icons/io"
// import * as Ionicons5 from "react-icons/io5"
// import * as MaterialDesignIcons from "react-icons/md"
// import * as RemixIcon from "react-icons/ri"
// import * as SimpleIcons from "react-icons/si"
// import * as Typicons from "react-icons/ti"
// import * as VSCodeIcons from "react-icons/vsc"
import * as WeatherIcons from "react-icons/wi"
// import * as CSS_gg from "react-icons/cg"
// import * as PfIcons from "@patternfly/react-icons"
import { IconType } from "../../../Constants/Widgets/ImagePanel"
import { getValue } from "../../../Util/Util"

const icons = {
  // ai: AntDesignIcons,
  // bs: BootstrapIcons,
  // bi: BoxIcons,
  // di: Devicons,
  // fi: Feather,
  // fc: FlatColorIcons,
  // fa: FontAwesome,
  // gi: GameIcons,
  // go: GithubOcticonsIcons,
  // gr: GrommetIcons,
  // hi: Heroicons,
  // im: IcoMoonFree,
  // io: Ionicons4,
  // io5: Ionicons5,
  // md: MaterialDesignIcons,
  // ri: RemixIcon,
  // si: SimpleIcons,
  // ti: Typicons,
  // vsc: VSCodeIcons,
  wi: WeatherIcons,
  // cg: CSS_gg,
  // pf: PfIcons,
}

// commented out most of the icons pack
// * to reduce the overall build size
// * to avoid out of memory in node build


// output: FcEmptyBattery => ['Fc', 'Empty', 'Battery']
const iconNameRegex = /[A-Z][a-z]+|[0-9]+/g

export const getIcon = (iconName = "", rotation = 0, dimensions = {}) => {
  const rawString = iconName.match(iconNameRegex)
  let iconType = IconType.PatternFlyIcons
  if (rawString !== null && rawString.length > 1) {
    iconType = rawString[0].toLocaleLowerCase()
  }

  // pattern fly icon name not starts with Pf, remove the Pf from the icon name
  if (iconType === IconType.PatternFlyIcons) {
    iconName = rawString.slice(1).join("")
  }

  let iconsObject = getValue(icons, iconType, null)

  const FinalIcon = getValue(iconsObject, iconName, QuestionIcon)
  return (
    <FinalIcon style={{ width: dimensions.width, height: "100%", transform: `rotate(${rotation}deg)` }} />
  )
}
