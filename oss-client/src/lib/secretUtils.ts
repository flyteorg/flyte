export const getSecretScope = (domain?: string, project?: string) => {
  const domainLength = (domain ?? '').length
  const projectLength = (project ?? '').length
  const domainLabel =
    String(domain).charAt(0).toUpperCase() + String(domain).slice(1)

  if (domainLength === 0 && projectLength === 0) {
    return { title: 'Full organization', subtitle: 'All projects and domains' }
  }

  if (domainLength > 0 && projectLength === 0) {
    return { title: 'Domain', subtitle: `All projects in ${domainLabel}` }
  }

  if (domain && domainLength > 0 && projectLength > 0) {
    return {
      title: 'Project + Domain',
      subtitle: `Only ${project} / ${domainLabel}`,
    }
  }

  return { title: '-', subtitle: '' }
}
