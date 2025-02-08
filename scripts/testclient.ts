import { createApiClient } from './apiclient'

const apiclient = createApiClient('http://localhost:8080', {
  beforeRequest(config) {
    if (config?.headers) {
      config.headers['Authorization'] = 'lol'
    }
  },
})

console.log('HelloWorld:', await apiclient.main.HelloWorld())
console.log(
  'ExampleHandler1:',
  await apiclient.main.ExampleHandler1({
    name: 'name',
    users: [{ name: 'name', age: 0 }],
  }),
)
console.log(
  'ExampleHandler2:',
  await apiclient.main.ExampleHandler2({
    name: 'name',
    users: [{ name: 'name', age: 0 }],
  }),
)
console.log('SomeHandler:', await apiclient.pkg.SomeHandler())
console.log('getPlaylists:', await apiclient.main.getPlaylists())
console.log('Server_theHandler:', await apiclient.main.Server_theHandler())
console.log(
  'Server_theHandlerPtr:',
  await apiclient.main.Server_theHandlerPtr(),
)
