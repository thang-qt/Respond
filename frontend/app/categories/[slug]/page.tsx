import { redirect } from "next/navigation"

export default async function CategorySlugRedirectPage({
  params,
  searchParams,
}: {
  params: Promise<{ slug: string }>
  searchParams?: Promise<Record<string, string | string[] | undefined>>
}) {
  const { slug } = await params
  const queryObject = (await searchParams) ?? {}
  const query = new URLSearchParams()

  for (const [key, value] of Object.entries(queryObject)) {
    if (Array.isArray(value)) {
      value.forEach((entry) => query.append(key, entry))
      continue
    }
    if (typeof value === "string") {
      query.set(key, value)
    }
  }

  const nextPath = `/tags/${encodeURIComponent(slug)}`
  redirect(query.size > 0 ? `${nextPath}?${query.toString()}` : nextPath)
}
