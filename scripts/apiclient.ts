export interface ApiError {
  message: string
  statusCode?: number
}
export type ApiResponse<T> =
  | { data: T; error: null }
  | { data: null; error: ApiError }

export type mainExampleHandler1Input = {
  name: string
  users: {
    name: string
    age: number
  }[]
}

export type mainExampleHandler1Output = {
  greeting: string
}

type mainExampleHandler1Handler = (
  params: mainExampleHandler1Input
) => Promise<ApiResponse<mainExampleHandler1Output>>

export type mainExampleHandler2Input = {
  name: string
  users: {
    name: string
    age: number
  }[]
}

export type mainExampleHandler2Output = {
  greeting: string
}

type mainExampleHandler2Handler = (
  params: mainExampleHandler2Input
) => Promise<ApiResponse<mainExampleHandler2Output>>

export type mainHelloWorldInput = {}

export type mainHelloWorldOutput = string

type mainHelloWorldHandler = (
  params: mainHelloWorldInput
) => Promise<ApiResponse<mainHelloWorldOutput>>

export type maingetPlaylistsInput = {}

export type maingetPlaylistsOutput = {
  id?: string
  playlistId?: string
  title?: string
  pinned?: boolean
  description?: string
}[]

type maingetPlaylistsHandler = (
  params: maingetPlaylistsInput
) => Promise<ApiResponse<maingetPlaylistsOutput>>

export type pkgSomeHandlerInput = {}

export type pkgSomeHandlerOutput = string

type pkgSomeHandlerHandler = (
  params: pkgSomeHandlerInput
) => Promise<ApiResponse<pkgSomeHandlerOutput>>

export interface ApiClient {
  main: {
    ExampleHandler1: mainExampleHandler1Handler
    ExampleHandler2: mainExampleHandler2Handler
    HelloWorld: mainHelloWorldHandler
    getPlaylists: maingetPlaylistsHandler
  }
  pkg: {
    SomeHandler: pkgSomeHandlerHandler
  }
}

type ApiClientConfig = {
  beforeRequest?: (config: RequestInit) => void | Promise<void>
}

export function createApiClient(
  baseUrl: string,
  config?: ApiClientConfig
): ApiClient {
  async function doFetch(path: string, params: unknown) {
    try {
      const requestConfig: RequestInit = {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(params),
      }

      if (config?.beforeRequest) {
        await config.beforeRequest(requestConfig)
      }

      const response = await fetch(`${baseUrl}/${path}`, requestConfig)
      if (!response.ok) {
        return {
          data: null,
          error: {
            message: "API request failed",
            statusCode: response.status,
          },
        }
      }
      const data = await response.json()
      return { data, error: null }
    } catch (error) {
      return {
        data: null,
        error: {
          message:
            error instanceof Error ? error.message : "Unknown error occurred",
        },
      }
    }
  }
  const client: ApiClient = {
    main: {
      ExampleHandler1: (params) => doFetch("main.ExampleHandler1", params),
      ExampleHandler2: (params) => doFetch("main.ExampleHandler2", params),
      HelloWorld: (params) => doFetch("main.HelloWorld", params),
      getPlaylists: (params) => doFetch("main.getPlaylists", params),
    },
    pkg: {
      SomeHandler: (params) => doFetch("pkg.SomeHandler", params),
    },
  }
  return client
}
