export * from './collections.service'
import { CollectionsService } from './collections.service'
export * from './common.service'
import { CommonService } from './common.service'
export * from './features.service'
import { FeaturesService } from './features.service'
export const APIS = [CollectionsService, CommonService, FeaturesService]
