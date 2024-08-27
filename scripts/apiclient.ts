export interface ApiError {
  message: string
  statusCode?: number
}
export type ApiResponse<T> =
  | { data: T; error: null }
  | { data: null; error: ApiError }
export interface ApiClient {
  pkg: {
    SomeHandler: (params: {}) => Promise<ApiResponse<string>>
  }
  main: {
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
    HelloWorld: (params: {}) => Promise<ApiResponse<string>>
  }
  beforeRequest(hook: (config: RequestInit) => void): void
}

export const createApiClient = (baseUrl: string): ApiClient => {
  let beforeRequestHook: ((config: RequestInit) => void) | null = null

  async function doFetch(path: string, params: unknown) {
    try {
      const config: RequestInit = {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(params),
      }

      if (beforeRequestHook) {
        beforeRequestHook(config)
      }

      const response = await fetch(`${baseUrl}/${path}`, config)
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
    pkg: {
      SomeHandler: (params) => doFetch("pkg.SomeHandler", params),
    },
    main: {
      ExampleHandler2: (params) => doFetch("main.ExampleHandler2", params),
      HelloWorld: (params) => doFetch("main.HelloWorld", params),
      getPlaylists: (params) => doFetch("main.getPlaylists", params),
      ExampleHandler1: (params) => doFetch("main.ExampleHandler1", params),
    },
    beforeRequest: (hook) => {
      beforeRequestHook = hook
    },
  }
  return client
}
