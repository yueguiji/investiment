export function splitTagString(value) {
  return String(value || '')
    .split(/[,\n]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

export function joinTagArray(values) {
  return Array.from(new Set((values || []).map((item) => String(item).trim()).filter(Boolean))).join(',')
}

export function buildTagOptions(groups) {
  return (groups || []).flatMap((group) =>
    (group.options || []).map((option) => ({
      label: option.label,
      value: option.value,
      group: group.label
    }))
  )
}

export function getTagLabel(groups, value) {
  for (const group of groups || []) {
    for (const option of group.options || []) {
      if (option.value === value) {
        return option.label
      }
    }
  }
  return value
}

export function collectTemplateTags(template) {
  return [
    ...splitTagString(template.styleTags),
    ...splitTagString(template.emotionTags),
    ...splitTagString(template.volumeTags),
    ...splitTagString(template.scenarioTags),
    ...splitTagString(template.capitalTags),
    ...splitTagString(template.factorTags)
  ]
}
