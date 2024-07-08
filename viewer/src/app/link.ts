export type Link = {
  /**
   * Supplies the URI to a remote resource (or resource fragment).
   */
  href: string
  /**
   * A hint indicating what the language of the result of dereferencing the link should be.
   */
  hreflang?: string
  length?: number
  /**
   * The type or semantics of the relation.
   */
  rel: string
  /**
   * Use `true` if the `href` property contains a URI template with variables that needs to be substituted by values to get a URI
   */
  templated?: boolean
  /**
   * Used to label the destination of a link such that it can be used as a human-readable identifier.
   */
  title?: string
  /**
   * A hint indicating what the media type of the result of dereferencing the link should be.
   */
  type?: string
  /**
   * Without this parameter you should repeat a link for each media type the resource is offered.
   * Adding this parameter allows listing alternative media types that you can use for this resource. The value in the `type` parameter becomes the recommended media type.
   */
  types?: Array<string>
  /**
   * A base path to retrieve semantic information about the variables used in URL template.
   */
  varBase?: string
}
