import { createApiClient, main_PointersAreUndefined } from './apiclient'

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
    users: [{ name: 'name', age: 0, created: '' }],
  }),
)
console.log(
  'ExampleHandler2:',
  await apiclient.main.ExampleHandler2({
    name: 'name',
    users: [{ name: 'name', age: 0, created: '' }],
  }),
)
console.log('SomeHandler:', await apiclient.pkg.SomeHandler())
console.log('getPlaylists:', await apiclient.main.getPlaylists())
console.log('theHandler:', await apiclient.main.theHandler())
console.log('theHandlerPtr:', await apiclient.main.theHandlerPtr())

let ptrs: main_PointersAreUndefined = {}
console.log('weHandleInputPointers:', await apiclient.main.weHandleInputPointers(ptrs))

console.log(
  'weAlsoHandleEnums opt 1 result',
  await apiclient.main.weAlsoHandleEnums({ Opt1: 'hello' }),
)

console.log(
  'weAlsoHandleEnums opt 2 result',
  await apiclient.main.weAlsoHandleEnums({
    Opt2: { name: 'john salchichon', age: 28 },
  }),
)
