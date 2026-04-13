export async function logBackendCall<T>(name: string, args: unknown[], call: () => Promise<T>): Promise<T> {
  console.log(`[backend->frontend] ${name}:request`, ...args)
  try {
    const result = await call()
    console.log(`[backend->frontend] ${name}:response`, result)
    return result
  } catch (error) {
    console.error(`[backend->frontend] ${name}:error`, error)
    throw error
  }
}
