interface ApiOptions extends RequestInit {
  params?: Record<string, string>;
}

/**
 * API utility function that automatically includes credentials
 * and handles common response patterns
 */
export async function api<T = any>(
  endpoint: string, 
  options: ApiOptions = {}
): Promise<T> {
  const { params, ...fetchOptions } = options;
  
  // Build URL with query parameters if provided
  const url = new URL(endpoint, window.location.origin);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      url.searchParams.append(key, value);
    });
  }
  
  // Set default options for all requests
  const defaultOptions: RequestInit = {
    credentials: 'include',  // Always send cookies
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  };
  
  // Merge options
  const finalOptions = { ...defaultOptions, ...fetchOptions };
  
  // Make the request
  const response = await fetch(url.toString(), finalOptions);
  
  // Handle errors
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(
      errorData.message || 
      `API request failed with status ${response.status}`
    );
  }
  
  // Return JSON response, or an empty object for 204 No Content
  if (response.status === 204) {
    return {} as T;
  }
  
  return response.json();
}

// Convenience methods
export const apiGet = <T = any>(endpoint: string, options?: ApiOptions) => 
  api<T>(endpoint, { method: 'GET', ...options });

export const apiPost = <T = any>(endpoint: string, data?: any, options?: ApiOptions) => 
  api<T>(endpoint, { 
    method: 'POST',
    body: data ? JSON.stringify(data) : undefined,
    ...options 
  });

export const apiPut = <T = any>(endpoint: string, data?: any, options?: ApiOptions) => 
  api<T>(endpoint, { 
    method: 'PUT',
    body: data ? JSON.stringify(data) : undefined,
    ...options 
  });

export const apiDelete = <T = any>(endpoint: string, data?: any, options?: ApiOptions) => 
  api<T>(endpoint, { 
    method: 'DELETE',
    body: data ? JSON.stringify(data) : undefined,
    ...options 
  });
