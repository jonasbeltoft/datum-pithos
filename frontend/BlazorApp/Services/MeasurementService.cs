using System;
using System.ComponentModel.DataAnnotations;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using System.Text.Json.Serialization;
using BlazorApp.Components.Pages.Measurements;

namespace BlazorApp.Services;

public class MeasurementService
{
	private readonly HttpClient httpClient;
	public Collection[] collections = Array.Empty<Collection>();
	public Unit[] units = Array.Empty<Unit>();

	public bool IsLoaded { get; set; } = false;

	public MeasurementService(HttpClient _httpClient)
	{
		httpClient = _httpClient;
	}

	// <summary>
	// Fetches measurement data from the API.
	// </summary>
	public async Task<MeasurementCollection?> FetchMeasurementsAsync(int CollectionId)
	{
		// collection_id: int,
		// sample_id: int,
		// page_size: int,
		// page: int,
		// before: int, // UNIX timestamp in seconds
		// after: int // UNIX timestamp in seconds

		try
		{
			var response = await httpClient.GetAsync($"samples?collection_id={CollectionId}");
			if (response == null)
			{
				Console.WriteLine("Failed to fetch measurements: response is null");
				return null;
			}

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch measurements: {response.StatusCode}");
				return null;
			}

			string jsonString = await response.Content.ReadAsStringAsync();
			if (string.IsNullOrEmpty(jsonString))
			{
				Console.WriteLine("Failed to fetch measurements: response content is empty");
				return null;
			}

			var measurements = JsonSerializer.Deserialize<MeasurementCollection>(jsonString);
			if (measurements == null)
			{
				Console.WriteLine("Failed to deserialize measurements: response content is not valid JSON");
				return null;
			}

			if (units.Length == 0)
			{
				units = await FetchUnitsAsync();
			}
			// Insert the unit name into the headers
			foreach (var header in measurements.headers)
			{
				var unit = units.FirstOrDefault(u => u.Id == header.unitId);
				if (unit != null)
				{
					header.unit = unit.Name;
				}
			}

			return measurements;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching measurements: {ex.Message}");
			return null;
		}
	}

	// <summary>
	// Deletes a measurement by its ID.
	// </summary>
	public async Task<bool> DeleteMeasurementAsync(int sampleId)
	{
		try
		{
			HttpResponseMessage response = await httpClient.DeleteAsync($"samples?sample_id={sampleId}");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to delete measurement: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error deleting measurement: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Updates a measurement/sample.
	// </summary>
	public async Task<bool> UpdateMeasurementAsync(int sampleId, long? created_at, string? note)
	{
		// Query params:
		//	 sample_id: int,
		//	 created_at: int, // UNIX timestamp in seconds
		//	 note: string,
		if (created_at == null && note == null)
		{
			return false;
		}

		string qparams = "";
		if (created_at != null)
		{
			qparams += $"&created_at={created_at}";
		}
		if (note != null)
		{
			qparams += $"&note={Uri.EscapeDataString(note)}";
		}

		try
		{
			HttpResponseMessage response = await httpClient.PutAsync($"samples?sample_id={sampleId}{qparams}", new StringContent(""));

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to update measurement: {response.StatusCode} {response.Content.ReadAsStringAsync().Result}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error updating measurement: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Updates a value in a measurement.
	// </summary>
	public async Task<bool> UpdateMeasurementValueAsync(int sampleId, int attributeId, string value)
	{
		try
		{
			HttpResponseMessage response = await httpClient.PostAsync($"sample-values?sample_id={sampleId}&attribute_id={attributeId}&value={Uri.EscapeDataString(value)}", new StringContent(""));

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to update measurement: {response.StatusCode} {response.Content.ReadAsStringAsync().Result}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error updating measurement: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Creates a new collection of measurements.
	// </summary>
	public async Task<(bool, string)> CreateMeasurementCollectionAsync(string name, string? description = "")
	{
		var content = new
		{
			name,
			description,
		};
		try
		{
			HttpResponseMessage response = await httpClient.PostAsJsonAsync("collections", content);

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to create measurement collection: {response.StatusCode}");
				return (false, response.Content.ReadAsStringAsync().Result);
			}
			return (true, "");
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error creating measurement collection: {ex.Message}");
			return (false, ex.Message);
		}
	}

	// <summary>
	// Fetches all collections
	// </summary>
	public async Task<Collection[]> FetchCollectionsAsync()
	{
		try
		{
			HttpResponseMessage response = await httpClient.GetAsync("collections");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch collections: {response.StatusCode}");
				IsLoaded = true;
				return Array.Empty<Collection>();
			}

			string jsonString = await response.Content.ReadAsStringAsync();

			collections = JsonSerializer.Deserialize<Collection[]>(jsonString) ?? Array.Empty<Collection>();
			IsLoaded = true;
			return collections;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching collections: {ex.Message}");
			IsLoaded = true;
			return Array.Empty<Collection>();
		}
	}

	// <summary>
	// Deletes a collection by its ID.
	// </summary>
	public async Task<bool> DeleteCollectionAsync(int id)
	{
		try
		{
			HttpResponseMessage response = await httpClient.DeleteAsync($"collections?collection_id={id}");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to delete collection: {response.StatusCode}");
			}
			// Delete the collection from the local array
			collections = collections.Where(c => c.Id != id).ToArray();
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error deleting collection: {ex.Message}");
		}
		return false;
	}

	// <summary>
	// Fetches logs.
	// </summary>
	public async Task<LogEntry[]> FetchLogsAsync(int? userId, string? httpMethod)
	{
		// Parameters:
		// http_method: string,
		// user_id: int,
		try
		{
			var args = new Dictionary<string, string>();
			if (userId != null)
			{
				args.Add("user_id", userId.ToString()!);
			}
			if (httpMethod != null)
			{
				args.Add("http_method", httpMethod);
			}
			var queryString = string.Join("&", args.Select(kvp => $"{kvp.Key}={Uri.EscapeDataString(kvp.Value)}"));
			if (!string.IsNullOrEmpty(queryString))
			{
				queryString = "?" + queryString;
			}
			HttpResponseMessage response = await httpClient.GetAsync($"logs{queryString}");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch logs: {response.StatusCode}");
				return [];
			}

			string jsonString = await response.Content.ReadAsStringAsync();
			if (string.IsNullOrEmpty(jsonString))
			{
				Console.WriteLine("Failed to fetch logs: response content is empty");
				return [];
			}
			var logs = JsonSerializer.Deserialize<LogEntry[]>(jsonString);

			return logs ?? [];
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching logs: {ex.Message}");
			return [];
		}
	}

	// <summary>
	// Adds a new measurement to a collection.
	// </summary>
	public async Task<bool> AddMeasurementAsync(int collectionId)
	{
		// 	{
		// 	collection_id: int,
		// 	note: string,
		// 	values: [
		// 		{
		// 			attribute_id: int,
		// 			value: string
		// 		}
		// 		...
		// 	]
		// }
		var content = new
		{
			collection_id = collectionId,
			values = Array.Empty<object>(),
		};
		try
		{
			HttpResponseMessage response = await httpClient.PostAsJsonAsync("samples", content);

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to add measurement: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error adding measurement: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Fetches all units.
	// </summary>
	public async Task<Unit[]> FetchUnitsAsync()
	{
		try
		{
			HttpResponseMessage response = await httpClient.GetAsync("units");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch units: {response.StatusCode}");
				return Array.Empty<Unit>();
			}

			string jsonString = await response.Content.ReadAsStringAsync();

			units = JsonSerializer.Deserialize<Unit[]>(jsonString) ?? Array.Empty<Unit>();
			return units;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching units: {ex.Message}");
			return Array.Empty<Unit>();
		}
	}

	// <summary>
	// Add new attribute (column) to a collection.
	// </summary>
	public async Task<bool> AddAttributeAsync(int collectionId, string name, int? unitId)
	{

		// Add variables to query params
		var args = new Dictionary<string, string>
		{
			{ "collection_id", collectionId.ToString() },
			{ "name", name },
		};
		if (unitId != null)
		{
			args.Add("unit_id", unitId.ToString()!);
		}
		var queryString = string.Join("&", args.Select(kvp => $"{kvp.Key}={Uri.EscapeDataString(kvp.Value)}"));
		if (!string.IsNullOrEmpty(queryString))
		{
			queryString = "?" + queryString;
		}

		try
		{
			HttpResponseMessage response = await httpClient.PostAsync($"attributes{queryString}", new StringContent(""));

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to add attribute: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error adding attribute: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Deletes an attribute (column) from a collection.
	// </summary>
	public async Task<bool> DeleteAttributeAsync(int attributeId)
	{
		// attribute_id: int
		try
		{
			HttpResponseMessage response = await httpClient.DeleteAsync($"attributes?attribute_id={attributeId}");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to delete attribute: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error deleting attribute: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Add User
	// </summary>
	public async Task<bool> AddUserAsync(User user)
	{
		try
		{
			var paramsDict = new Dictionary<string, string>
			{
				{ "username", user.Username },
				{ "password", user.Password },
				{ "role_id", user.RoleId.ToString() }
			};

			var qparams = string.Join("&", paramsDict.Select(kvp => $"{kvp.Key}={Uri.EscapeDataString(kvp.Value)}"));
			if (!string.IsNullOrEmpty(qparams))
			{
				qparams = "?" + qparams;
			}
			else
			{
				return false;
			}

			HttpResponseMessage response = await httpClient.PostAsync($"users{qparams}", new StringContent(""));

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to add user: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error adding user: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Fetches all users.
	// </summary>
	public async Task<User[]?> FetchUsersAsync()
	{
		try
		{
			HttpResponseMessage response = await httpClient.GetAsync("users");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch users: {response.StatusCode}");
				return null;
			}

			string jsonString = await response.Content.ReadAsStringAsync();

			var users = JsonSerializer.Deserialize<User[]>(jsonString) ?? null;
			return users;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching users: {ex.Message}");
			return null;
		}
	}

	// <summary>
	// Deletes a user by its ID.
	// </summary>
	public async Task<bool> DeleteUserAsync(int id)
	{
		try
		{
			HttpResponseMessage response = await httpClient.DeleteAsync($"users?id={id}");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to delete user: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error deleting user: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Get all roles.
	// </summary>
	public async Task<Role[]> FetchRolesAsync()
	{
		try
		{
			HttpResponseMessage response = await httpClient.GetAsync("roles");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch roles: {response.StatusCode}");
				return Array.Empty<Role>();
			}

			string jsonString = await response.Content.ReadAsStringAsync();

			var roles = JsonSerializer.Deserialize<Role[]>(jsonString) ?? [];
			return roles;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching roles: {ex.Message}");
			return Array.Empty<Role>();
		}
	}

	// <summary>
	// Add a new unit.
	// </summary>
	public async Task<bool> AddUnitAsync(string name)
	{
		try
		{
			HttpResponseMessage response = await httpClient.PostAsync($"units?name={Uri.EscapeDataString(name)}", new StringContent(""));

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to add unit: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error adding unit: {ex.Message}");
			return false;
		}
	}

	public int[] GetCollectionIds()
	{
		return collections.Select(c => c.Id).ToArray();
	}

	// <summary>
	// Gets a collection by its ID with an ok value if none was found.
	// </summary>
	public (Collection, bool) GetCollectionById(int id)
	{
		Collection collection;
		try
		{
			if (collections.Length == 0)
			{
				return (new Collection(), false);
			}
			collection = collections.First(c => c.Id == id);
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error getting collection by ID: {ex.Message}");
			return (new Collection(), false);
		}
		return (collection, true);
	}
}

public class Role
{
	[JsonPropertyName("id")]
	public int Id { get; set; }

	[JsonPropertyName("name")]
	public string Name { get; set; } = string.Empty;
}

public class User
{
	[JsonPropertyName("id")]
	public int Id { get; set; }

	[Required(ErrorMessage = "Brugernavn er påkrævet")]
	[MaxLength(50, ErrorMessage = "Maksimal længde er 50 tegn")]
	[MinLength(8, ErrorMessage = "Minimum længde er 8 tegn")]
	[JsonPropertyName("username")]
	public string Username { get; set; } = string.Empty;

	[Required(ErrorMessage = "Password er påkrævet")]
	[MaxLength(50, ErrorMessage = "Maksimal længde er 50 tegn")]
	[MinLength(8, ErrorMessage = "Minimum længde er 8 tegn")]
	[JsonPropertyName("password")]
	public string Password { get; set; } = string.Empty;

	[JsonPropertyName("displayName")]
	public string DisplayName { get; set; } = string.Empty;

	// 0 is default, so check if it is set to 0 or not
	[Required(ErrorMessage = "Rolle er påkrævet")]
	[Range(1, int.MaxValue, ErrorMessage = "Rolle er påkrævet")]
	[JsonPropertyName("role_id")]
	public int RoleId { get; set; }
}

public class LogEntry
{
	public int Id { get; set; }

	[JsonPropertyName("created_at")]
	public long CreatedAt { get; set; }

	[JsonPropertyName("instance_user")]
	public int InstanceUserId { get; set; }

	[JsonPropertyName("instance_username")]
	public string InstanceUsername { get; set; } = string.Empty;

	[JsonPropertyName("crud_action")]
	public string CrudAction { get; set; } = string.Empty;

	[JsonPropertyName("request_url")]
	public string RequestUrl { get; set; } = string.Empty;

	[JsonPropertyName("request_body")]
	public string RequestBody { get; set; } = string.Empty;

	[JsonPropertyName("response_code")]
	public int ResponseCode { get; set; }
}


public class Unit
{
	[JsonPropertyName("id")]
	public int Id { get; set; }

	[JsonPropertyName("name")]
	public string Name { get; set; } = "";
}

public class Collection
{
	[JsonPropertyName("id")]
	public int Id { get; set; }

	[JsonPropertyName("name")]
	public string Name { get; set; } = "";

	[JsonPropertyName("description")]
	public string Description { get; set; } = "";
}

public class MeasurementCollection
{
	[JsonPropertyName("attributes")]
	public List<MeasurementHeader> headers { get; set; } = new();

	[JsonPropertyName("samples")]
	public List<Measurement> entries { get; set; } = new();

	[JsonPropertyName("total_count")]
	public int TotalCount { get; set; }

	[JsonPropertyName("page_info")]
	public PageInfo PageInfo { get; set; } = new();
}

public class MeasurementHeader
{
	[JsonPropertyName("attribute_id")]
	public int Id { get; set; }

	public string name { get; set; } = string.Empty;

	[JsonPropertyName("unit_id")]
	public int unitId { get; set; }

	public string unit { get; set; } = string.Empty;
}
public class Measurement
{
	[JsonPropertyName("sample_id")]
	public int Id { get; set; }

	[JsonPropertyName("note")]
	public string Note { get; set; } = string.Empty;

	[JsonPropertyName("created_at")]
	public long CreatedAt { get; set; }

	public List<MeasurementValue> values { get; set; } = new();
}
public class MeasurementValue
{
	[JsonPropertyName("attribute_id")]
	public int AttributeId { get; set; }

	public string value { get; set; } = string.Empty;
}
public class PageInfo
{
	public int Limit { get; set; }
	public int Offset { get; set; }

	[JsonPropertyName("has_next_page")]
	public bool HasNextPage { get; set; }
}