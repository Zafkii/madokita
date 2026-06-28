import type { PlatformAPI } from "../platform/shared/PlatformAPI"

export interface RuntimeSystem {
  readonly id: string
  initialize?(platform: PlatformAPI): Promise<void>
  start?(platform: PlatformAPI): Promise<void>
  ready?(platform: PlatformAPI): Promise<void>
  shutdown?(platform: PlatformAPI): Promise<void>
}
