import { Pipe, PipeTransform } from '@angular/core'

@Pipe({
  name: 'propertyValue',
  standalone: true,
})
export class PropertyValuePipe implements PipeTransform {
  transform(properties: unknown, key: string): string {
    const p = properties as { [key: string]: unknown }
    return (p[key] as string) || ''
  }
}
