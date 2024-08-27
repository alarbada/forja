export interface ApiError {
  message: string
  statusCode?: number
}
export type ApiResponse<T> =
  | { data: T; error: null }
  | { data: null; error: ApiError }
export interface ApiClient {
  main: {
    HelloWorld: (params: {}) => Promise<ApiResponse<string>>
    getPlaylists: (params: {}) => Promise<
      ApiResponse<
        {
          id?: string
          playlistId?: string
          title?: string
          pinned?: boolean
          description?: string
        }[]
      >
    >
    ExampleHandler1: (params: {
      name: string
      users: {
        name: string
        age: number
      }[]
    }) => Promise<
      ApiResponse<{
        greeting: string
      }>
    >
    ExampleHandler2: (params: {
      name: string
      users: {
        name: string
        age: number
      }[]
    }) => Promise<
      ApiResponse<{
        greeting: string
      }>
    >
  }
  pkg: {
    SomeHandler: (params: {}) => Promise<ApiResponse<string>>
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
      ExampleHandler2: (params) => doFetch("main.ExampleHandler2", params),
      HelloWorld: (params) => doFetch("main.HelloWorld", params),
      getPlaylists: (params) => doFetch("main.getPlaylists", params),
      ExampleHandler1: (params) => doFetch("main.ExampleHandler1", params),
    },
    pkg: {
      SomeHandler: (params) => doFetch("pkg.SomeHandler", params),
    },
  }
  return client
}
