export const getDynamicFilter = (key, value, existingFilter = []) => {
  const finalFilters = []
  if (value && value.length > 3 && value.includes("=", 2)) {
    const rawFilters = []
    if (value.includes(",")) {
      rawFilters.push(...value.split(","))
    } else {
      rawFilters.push(value)
    }
    for (let index = 0; index < rawFilters.length; index++) {
      const kv = rawFilters[index].split("=", 2)
      if (kv.length !== 2) {
        continue
      }
      finalFilters.push({ k: kv[0], o: "regex", v: kv[1] })
    }
  } else {
    finalFilters.push({ k: key, o: "regex", v: value })
  }
  // add if additional filters supplied
  finalFilters.push(...existingFilter)

  //console.log(JSON.stringify(finalFilters))

  return finalFilters
}
