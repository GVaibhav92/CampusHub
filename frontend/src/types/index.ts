export type Role = 'STUDENT' | 'FACULTY' | 'ADMIN' | 'TECHNICIAN'

export interface Tenant {
  id: string
  name: string
  slug: string
}

export interface User {
  id: string
  email: string
  role: Role
  tenant_id: string
}

export type ResourceStatus = 'AVAILABLE' | 'MAINTENANCE'

export interface Resource {
  ID: string
  TenantID: string
  Name: string
  Type: string
  Capacity: number
  Location: string
  Status: ResourceStatus
  CreatedAt: string
}

export type BookingStatus = 'CONFIRMED' | 'CANCELLED'

export interface Booking {
  ID: string
  TenantID: string
  ResourceID: string
  UserID: string
  StartTime: string
  EndTime: string
  Status: BookingStatus
  CreatedAt: string
}

export type IssueStatus = 'OPEN' | 'IN_PROGRESS' | 'RESOLVED' | 'CLOSED'
export type IssuePriority = 'LOW' | 'MEDIUM' | 'HIGH'

export interface Issue {
  ID: string
  TenantID: string
  ResourceID: string
  ReportedBy: string
  AssignedTo: string | null
  Category: string
  Priority: IssuePriority
  Description: string
  Status: IssueStatus
  AttachmentKey: string | null
  CreatedAt: string
  UpdatedAt: string
}

export interface ApiError {
  error: string
  status?: number
}
