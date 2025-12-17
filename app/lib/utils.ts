export function getRequestPreview(request: any, maxLength: number = 80): string {
  try {
    if (request?.contents?.[0]?.parts?.[0]?.text) {
      const text = request.contents[0].parts[0].text;
      if (text.length > maxLength) {
        return text.substring(0, maxLength) + "...";
      }
      return text;
    }
  } catch {}
  return "No preview";
}
