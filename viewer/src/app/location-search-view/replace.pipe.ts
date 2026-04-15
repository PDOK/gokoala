import { Pipe, PipeTransform } from '@angular/core'

@Pipe({
  name: 'replace',
  standalone: true,
})
export class ReplacePipe implements PipeTransform {
  transform(str: string, lookup: string, replaceVal: string): unknown {
    return str.replaceAll(lookup, replaceVal)
  }
}
