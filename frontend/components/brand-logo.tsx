import Image from "next/image"
import { cn } from "@/lib/utils"

export default function BrandLogo({
  size = 20,
  className,
  alt = "Respond.im",
}: {
  size?: number
  className?: string
  alt?: string
}) {
  return (
    <Image
      src="/logo.svg"
      alt={alt}
      width={size}
      height={size}
      className={cn("shrink-0", className)}
      priority
    />
  )
}
