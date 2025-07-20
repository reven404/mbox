#ifndef CLOUDFILES_BRIDGE_H
#define CLOUDFILES_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

// Windows Cloud Files API Bridge for Smart Folders
// Requires Windows 10 version 1809 (build 17763) or later

// Cloud Folder Registration
typedef struct {
    wchar_t* folder_path;
    wchar_t* display_name; 
    wchar_t* server_uri;
    int sync_root_id;
    int registered;
} CloudFolder;

// File Placeholder Information
typedef struct {
    wchar_t* relative_path;
    unsigned long long file_size;
    unsigned long long file_id;
    unsigned long file_attributes;
    unsigned long long created_time;
    unsigned long long modified_time;
    int is_directory;
    int hydration_state;
} FilePlaceholder;

// Hydration Policy
typedef enum {
    HYDRATION_POLICY_FULL = 0,
    HYDRATION_POLICY_PROGRESSIVE = 1,
    HYDRATION_POLICY_ALWAYS_FULL = 2,
    HYDRATION_POLICY_PARTIAL = 3
} HydrationPolicy;

// Cloud Files Operations
int cf_register_sync_root(const wchar_t* sync_root_path, const wchar_t* display_name, 
                         const wchar_t* server_uri);
int cf_unregister_sync_root(const wchar_t* sync_root_path);

// File Placeholder Management
int cf_create_placeholders(const wchar_t* sync_root_path, FilePlaceholder* placeholders, 
                          int count);
int cf_hydrate_placeholder(const wchar_t* file_path, HydrationPolicy policy);
int cf_dehydrate_placeholder(const wchar_t* file_path);
int cf_update_placeholder(const wchar_t* file_path, FilePlaceholder* info);

// Sync Root Status
int cf_get_sync_root_status(const wchar_t* sync_root_path);
int cf_set_in_sync_state(const wchar_t* file_path, int in_sync);

// File Operations Callbacks
typedef struct {
    int (*on_fetch_data)(const wchar_t* file_path, unsigned long long offset, 
                        unsigned long buffer_size, void* buffer);
    int (*on_validate_data)(const wchar_t* file_path);
    int (*on_cancel_fetch)(const wchar_t* file_path);
    int (*on_notification)(const wchar_t* file_path, int notification_type);
} CloudFileCallbacks;

int cf_register_callbacks(CloudFileCallbacks* callbacks);

// Utility Functions
int cf_get_placeholder_state(const wchar_t* file_path);
int cf_is_cloud_files_supported(void);
wchar_t* cf_get_last_error_message(void);
void cf_free_wstring(wchar_t* str);

// File Attributes
#define CLOUD_FILE_ATTRIBUTE_PLACEHOLDER           0x00000001
#define CLOUD_FILE_ATTRIBUTE_SYNC_ROOT             0x00000002
#define CLOUD_FILE_ATTRIBUTE_PINNED                0x00000004
#define CLOUD_FILE_ATTRIBUTE_UNPINNED              0x00000008
#define CLOUD_FILE_ATTRIBUTE_IN_SYNC               0x00000010
#define CLOUD_FILE_ATTRIBUTE_NOT_IN_SYNC           0x00000020

// Hydration States
#define PLACEHOLDER_HYDRATION_PARTIAL              0
#define PLACEHOLDER_HYDRATION_FULL                 1
#define PLACEHOLDER_HYDRATION_NOT_HYDRATED         2

// Notification Types
#define CLOUD_FILE_NOTIFY_FILE_FETCHED             1
#define CLOUD_FILE_NOTIFY_FILE_CREATED             2
#define CLOUD_FILE_NOTIFY_FILE_DELETED             3
#define CLOUD_FILE_NOTIFY_FILE_RENAMED             4

#ifdef __cplusplus
}
#endif

#endif // CLOUDFILES_BRIDGE_H