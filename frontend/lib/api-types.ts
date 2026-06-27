export interface ApiResponse<T> {
  data: T
}

export interface ApiListResponse<T> {
  data: T[]
  meta?: {
    page: number
    per_page: number
    total: number
    total_pages: number
    unread_count?: number
  }
}
