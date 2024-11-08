import { createApiClient } from './apiclient'

const apiclient = createApiClient('http://localhost:8080', {
  beforeRequest(config) {
    if (config?.headers) {
      config.headers['Authorization'] = 'lol'
    }
  },
})

console.log(await apiclient.main.HelloWorld())
console.log(
  await apiclient.main.ExampleHandler1({
    name: 'name',
    users: [{ name: 'name', age: 0 }],
  }),
)
console.log(
  await apiclient.main.ExampleHandler2({
    name: 'name',
    users: [{ name: 'name', age: 0 }],
  }),
)
console.log(await apiclient.pkg.SomeHandler())
console.log(await apiclient.main.getPlaylists())
