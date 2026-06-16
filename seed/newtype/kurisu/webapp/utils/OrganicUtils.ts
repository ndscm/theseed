export const GetOrganic = (
  organic?: string,
): "carbon" | "silicon" | undefined => {
  if (organic === "carbon") {
    return "carbon"
  }
  if (organic === "silicon") {
    return "silicon"
  }
  return undefined
}

export default { GetOrganic }
