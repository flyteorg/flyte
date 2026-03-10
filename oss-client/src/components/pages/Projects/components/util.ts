const whiteSpaceReg = /\s/
const rfc1123ForbiddenCharsRegex = /[^a-z0-9-]/g

export const getProjectId = (projectName: string) => {
  const splitReg = projectName?.trim().split(whiteSpaceReg)
  return splitReg
    ?.filter(Boolean)
    ?.join('-')
    .toLowerCase()
    .replace(rfc1123ForbiddenCharsRegex, '-')
    .trim()
}
