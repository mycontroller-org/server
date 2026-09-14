// Schedule Frequency values
export const ScheduleFrequency = {
  Daily: "daily",
  Weekly: "weekly",
  Monthly: "monthly",
  OnDate: "on_date",
}

// Schedule Frequency options
export const ScheduleFrequencyOptions = [
  {
    value: ScheduleFrequency.Daily,
    label: "opts.schedule_frequency.label_daily",
    description: "opts.schedule_frequency.desc_daily",
  },
  {
    value: ScheduleFrequency.Weekly,
    label: "opts.schedule_frequency.label_weekly",
    description: "opts.schedule_frequency.desc_weekly",
  },
  {
    value: ScheduleFrequency.Monthly,
    label: "opts.schedule_frequency.label_monthly",
    description: "opts.schedule_frequency.desc_monthly",
  },
  {
    value: ScheduleFrequency.OnDate,
    label: "opts.schedule_frequency.label_on_date",
    description: "opts.schedule_frequency.desc_on_date",
  },
]

export const WeekDayOptions = [
  { value: "sun", label: "opts.week_day.label_sunday" },
  { value: "mon", label: "opts.week_day.label_monday" },
  { value: "tue", label: "opts.week_day.label_tuesday" },
  { value: "wed", label: "opts.week_day.label_wednesday" },
  { value: "thu", label: "opts.week_day.label_thursday" },
  { value: "fri", label: "opts.week_day.label_friday" },
  { value: "sat", label: "opts.week_day.label_saturday" },
]

// Schedule type values
export const ScheduleType = {
  Repeat: "repeat",
  Cron: "cron",
  Simple: "simple",
  Sunrise: "sunrise",
  Sunset: "sunset",
}

// Schedule type options
export const ScheduleTypeOptions = [
  {
    value: ScheduleType.Repeat,
    label: "opts.schedule_type.label_repeat",
    description: "opts.schedule_type.desc_repeat",
  },
  {
    value: ScheduleType.Cron,
    label: "opts.schedule_type.label_cron",
    description: "opts.schedule_type.desc_cron",
  },
  {
    value: ScheduleType.Simple,
    label: "opts.schedule_type.label_simple",
    description: "opts.schedule_type.desc_simple",
  },
  {
    value: ScheduleType.Sunrise,
    label: "opts.schedule_type.label_sunrise",
    description: "opts.schedule_type.desc_sunrise",
  },
  {
    value: ScheduleType.Sunset,
    label: "opts.schedule_type.label_sunset",
    description: "opts.schedule_type.desc_sunset",
  },
]

// CustomVariable Types
export const CustomVariableType = {
  None: "none",
  Javascript: "javascript",
  Webhook: "webhook",
}

// CustomVariable type options
export const CustomVariableTypeOptions = [
  { value: CustomVariableType.None, label: "none" },
  { value: CustomVariableType.Javascript, label: "javascript" },
  { value: CustomVariableType.Webhook, label: "webhook" },
]
