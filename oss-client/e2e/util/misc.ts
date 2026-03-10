export async function withRetry<T>(
  operation: () => Promise<T>,
  maxRetries = 3,
  operationName = 'operation',
): Promise<T> {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      if (attempt > 1) {
        console.log(`${operationName} attempt ${attempt}/${maxRetries}`)
      }
      const result = await operation()
      if (attempt > 1) {
        console.log(`${operationName} successful on attempt ${attempt}`)
      }
      return result
    } catch (error) {
      console.log(`${operationName} attempt ${attempt} failed`)

      if (attempt === maxRetries) {
        const errorMessage =
          error instanceof Error ? error.message : String(error)
        throw new Error(
          `${operationName} failed after ${maxRetries} attempts. Last error: ${errorMessage}`,
        )
      }

      await new Promise((resolve) => setTimeout(resolve, 1000)) // Wait before retry
    }
  }
  throw new Error('Unexpected end of retry loop') // TypeScript safety
}
