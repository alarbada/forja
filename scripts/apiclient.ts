export interface ApiError {
  message: string
  statusCode?: number
}
export type ApiResponse<T> =
  | { data: T; error: null }
  | { data: null; error: ApiError }

export type MainHelloWorldInput = {}

export type MainHelloWorldOutput = string

type MainHelloWorldHandler = (
  params: MainHelloWorldInput,
) => Promise<ApiResponse<MainHelloWorldOutput>>

export type MainGetPlaylistsInput = {}

export type MainGetPlaylistsOutput = {
  id?: string
  playlistId?: string
  title?: string
  pinned?: boolean
  description?: string
}[]

type MainGetPlaylistsHandler = (
  params: MainGetPlaylistsInput,
) => Promise<ApiResponse<MainGetPlaylistsOutput>>

export type MainExampleHandler1Input = {
  name: string
  users: {
    name: string
    age: number
  }[]
}

export type MainExampleHandler1Output = {
  greeting: string
}

type MainExampleHandler1Handler = (
  params: MainExampleHandler1Input,
) => Promise<ApiResponse<MainExampleHandler1Output>>

export type MainExampleHandler2Input = {
  name: string
  users: {
    name: string
    age: number
  }[]
}

export type MainExampleHandler2Output = {
  greeting: string
}

type MainExampleHandler2Handler = (
  params: MainExampleHandler2Input,
) => Promise<ApiResponse<MainExampleHandler2Output>>

export type PkgSomeHandlerInput = {}

export type PkgSomeHandlerOutput = string

type PkgSomeHandlerHandler = (
  params: PkgSomeHandlerInput,
) => Promise<ApiResponse<PkgSomeHandlerOutput>>

export interface ApiClient {
  main: {
    ExampleHandler1: MainExampleHandler1Handler
    ExampleHandler2: MainExampleHandler2Handler
    HelloWorld: MainHelloWorldHandler
    getPlaylists: MainGetPlaylistsHandler
  }
  pkg: {
    SomeHandler: PkgSomeHandlerHandler
  }
}

type ApiClientConfig = {
  beforeRequest?: (config: RequestInit) => void | Promise<void>
}

export function createApiClient(
  baseUrl: string,
  config?: ApiClientConfig,
): ApiClient {
  async function doFetch(path: string, params: unknown) {
    try {
      const requestConfig: RequestInit = {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(params),
      }

      if (config?.beforeRequest) {
        await config.beforeRequest(requestConfig)
      }

      const response = await fetch(`${baseUrl}/${path}`, requestConfig)
      if (!response.ok) {
        const data = await response.json()
        const message = data.message

        return {
          data: null,
          error: { message, statusCode: response.status },
        }
      }
      const data = await response.json()
      return { data, error: null }
    } catch (error) {
      return {
        data: null,
        error: {
          message:
            error instanceof Error ? error.message : 'Unknown error occurred',
        },
      }
    }
  }
  const client: ApiClient = {
    main: {
      getPlaylists: (params) => doFetch('main.getPlaylists', params),
      ExampleHandler1: (params) => doFetch('main.ExampleHandler1', params),
      ExampleHandler2: (params) => doFetch('main.ExampleHandler2', params),
      HelloWorld: (params) => doFetch('main.HelloWorld', params),
    },
    pkg: {
      SomeHandler: (params) => doFetch('pkg.SomeHandler', params),
    },
  }
  return client
}
