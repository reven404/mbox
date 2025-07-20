#include "cloudfiles_bridge.h"

#ifdef _WIN32

#include <windows.h>
#include <cfapi.h>
#include <winioctl.h>
#include <iostream>
#include <string>
#include <vector>
#include <memory>

// Global variables
static CloudFileCallbacks* g_callbacks = nullptr;
static std::wstring g_last_error;

// Convert const char* to wchar_t*
std::wstring string_to_wstring(const char* str) {
    if (!str) return L"";
    int size = MultiByteToWideChar(CP_UTF8, 0, str, -1, nullptr, 0);
    if (size == 0) return L"";
    
    std::vector<wchar_t> buffer(size);
    MultiByteToWideChar(CP_UTF8, 0, str, -1, buffer.data(), size);
    return std::wstring(buffer.data());
}

// Convert wchar_t* to char*
std::string wstring_to_string(const std::wstring& wstr) {
    if (wstr.empty()) return "";
    int size = WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), -1, nullptr, 0, nullptr, nullptr);
    if (size == 0) return "";
    
    std::vector<char> buffer(size);
    WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), -1, buffer.data(), size, nullptr, nullptr);
    return std::string(buffer.data());
}

// Set last error message
void set_last_error(const std::wstring& message) {
    g_last_error = message;
    std::wcout << L"Cloud Files Error: " << message << std::endl;
}

// Cloud Files Callback Functions
void CALLBACK CloudFileCallback(
    _In_ CONST CF_CALLBACK_INFO* CallbackInfo,
    _In_ CONST CF_CALLBACK_PARAMETERS* CallbackParameters
) {
    if (!g_callbacks) return;
    
    switch (CallbackInfo->CallbackType) {
        case CF_CALLBACK_TYPE_FETCH_DATA:
        {
            if (g_callbacks->on_fetch_data) {
                std::wstring file_path = CallbackInfo->VolumeDosName;
                file_path += CallbackInfo->NormalizedPath;
                
                auto& fetch_data = CallbackParameters->FetchData;
                g_callbacks->on_fetch_data(
                    file_path.c_str(),
                    fetch_data.RequiredFileOffset.QuadPart,
                    fetch_data.RequiredLength.QuadPart,
                    nullptr // Buffer would be handled differently in real implementation
                );
            }
            break;
        }
        case CF_CALLBACK_TYPE_VALIDATE_DATA:
        {
            if (g_callbacks->on_validate_data) {
                std::wstring file_path = CallbackInfo->VolumeDosName;
                file_path += CallbackInfo->NormalizedPath;
                g_callbacks->on_validate_data(file_path.c_str());
            }
            break;
        }
        case CF_CALLBACK_TYPE_CANCEL_FETCH_DATA:
        {
            if (g_callbacks->on_cancel_fetch) {
                std::wstring file_path = CallbackInfo->VolumeDosName;
                file_path += CallbackInfo->NormalizedPath;
                g_callbacks->on_cancel_fetch(file_path.c_str());
            }
            break;
        }
        default:
            break;
    }
}

// Register sync root with Cloud Files API
int cf_register_sync_root(const wchar_t* sync_root_path, const wchar_t* display_name, 
                         const wchar_t* server_uri) {
    if (!sync_root_path || !display_name) {
        set_last_error(L"Invalid parameters for sync root registration");
        return -1;
    }

    // Prepare sync root info
    CF_SYNC_ROOT_STANDARD_INFO sync_root_info = {};
    sync_root_info.SyncRootFileId.QuadPart = 0;
    
    // Set display name
    size_t display_name_len = wcslen(display_name);
    if (display_name_len >= sizeof(sync_root_info.DisplayNameNormalForm) / sizeof(wchar_t)) {
        set_last_error(L"Display name too long");
        return -2;
    }
    wcscpy_s(sync_root_info.DisplayNameNormalForm, display_name);

    // Set server URI if provided
    if (server_uri) {
        size_t server_uri_len = wcslen(server_uri);
        if (server_uri_len < sizeof(sync_root_info.ProviderId) / sizeof(wchar_t)) {
            wcscpy_s(sync_root_info.ProviderId, server_uri);
        }
    }

    // Register callbacks
    CF_CALLBACK_REGISTRATION callbacks[] = {
        { CF_CALLBACK_TYPE_FETCH_DATA, CloudFileCallback },
        { CF_CALLBACK_TYPE_VALIDATE_DATA, CloudFileCallback },
        { CF_CALLBACK_TYPE_CANCEL_FETCH_DATA, CloudFileCallback },
        CF_CALLBACK_REGISTRATION_END
    };

    // Register the sync root
    HRESULT hr = CfRegisterSyncRoot(
        sync_root_path,
        callbacks,
        &sync_root_info,
        CF_REGISTER_FLAG_NONE
    );

    if (FAILED(hr)) {
        set_last_error(L"Failed to register sync root. Error code: " + std::to_wstring(hr));
        return -3;
    }

    std::wcout << L"Successfully registered sync root: " << sync_root_path << std::endl;
    return 0;
}

// Unregister sync root
int cf_unregister_sync_root(const wchar_t* sync_root_path) {
    if (!sync_root_path) {
        set_last_error(L"Invalid sync root path");
        return -1;
    }

    HRESULT hr = CfUnregisterSyncRoot(sync_root_path);
    if (FAILED(hr)) {
        set_last_error(L"Failed to unregister sync root. Error code: " + std::to_wstring(hr));
        return -2;
    }

    std::wcout << L"Successfully unregistered sync root: " << sync_root_path << std::endl;
    return 0;
}

// Create file placeholders
int cf_create_placeholders(const wchar_t* sync_root_path, FilePlaceholder* placeholders, 
                          int count) {
    if (!sync_root_path || !placeholders || count <= 0) {
        set_last_error(L"Invalid parameters for creating placeholders");
        return -1;
    }

    std::vector<CF_PLACEHOLDER_CREATE_INFO> create_infos(count);
    
    for (int i = 0; i < count; i++) {
        auto& info = create_infos[i];
        auto& placeholder = placeholders[i];
        
        // Set basic file information
        info.FileIdentity.QuadPart = placeholder.file_id;
        info.FileIdentityLength = sizeof(LARGE_INTEGER);
        
        info.RelativeFileName = placeholder.relative_path;
        info.FsMetadata.BasicInfo.FileSize.QuadPart = placeholder.file_size;
        info.FsMetadata.BasicInfo.FileAttributes = placeholder.file_attributes;
        
        // Convert timestamps
        info.FsMetadata.BasicInfo.CreationTime.QuadPart = placeholder.created_time;
        info.FsMetadata.BasicInfo.LastWriteTime.QuadPart = placeholder.modified_time;
        info.FsMetadata.BasicInfo.LastAccessTime.QuadPart = placeholder.modified_time;
        info.FsMetadata.BasicInfo.ChangeTime.QuadPart = placeholder.modified_time;

        info.Flags = CF_PLACEHOLDER_CREATE_FLAG_MARK_IN_SYNC;
    }

    HRESULT hr = CfCreatePlaceholders(
        sync_root_path,
        create_infos.data(),
        count,
        CF_CREATE_FLAG_NONE,
        nullptr
    );

    if (FAILED(hr)) {
        set_last_error(L"Failed to create placeholders. Error code: " + std::to_wstring(hr));
        return -2;
    }

    std::wcout << L"Successfully created " << count << L" placeholders" << std::endl;
    return 0;
}

// Hydrate placeholder file (download content)
int cf_hydrate_placeholder(const wchar_t* file_path, HydrationPolicy policy) {
    if (!file_path) {
        set_last_error(L"Invalid file path for hydration");
        return -1;
    }

    CF_HYDRATION_POLICY cf_policy;
    switch (policy) {
        case HYDRATION_POLICY_FULL:
            cf_policy = CF_HYDRATION_POLICY_FULL;
            break;
        case HYDRATION_POLICY_PROGRESSIVE:
            cf_policy = CF_HYDRATION_POLICY_PROGRESSIVE;
            break;
        case HYDRATION_POLICY_ALWAYS_FULL:
            cf_policy = CF_HYDRATION_POLICY_ALWAYS_FULL;
            break;
        default:
            cf_policy = CF_HYDRATION_POLICY_PARTIAL;
            break;
    }

    HRESULT hr = CfHydratePlaceholder(
        file_path,
        0, // Start offset
        MAXLONGLONG, // Length (full file)
        cf_policy,
        nullptr
    );

    if (FAILED(hr)) {
        set_last_error(L"Failed to hydrate placeholder. Error code: " + std::to_wstring(hr));
        return -2;
    }

    std::wcout << L"Successfully initiated hydration for: " << file_path << std::endl;
    return 0;
}

// Dehydrate placeholder file (remove content, keep metadata)
int cf_dehydrate_placeholder(const wchar_t* file_path) {
    if (!file_path) {
        set_last_error(L"Invalid file path for dehydration");
        return -1;
    }

    HRESULT hr = CfDehydratePlaceholder(
        file_path,
        0, // Start offset
        MAXLONGLONG, // Length (full file)
        CF_DEHYDRATE_FLAG_NONE,
        nullptr
    );

    if (FAILED(hr)) {
        set_last_error(L"Failed to dehydrate placeholder. Error code: " + std::to_wstring(hr));
        return -2;
    }

    std::wcout << L"Successfully dehydrated: " << file_path << std::endl;
    return 0;
}

// Update placeholder information
int cf_update_placeholder(const wchar_t* file_path, FilePlaceholder* info) {
    if (!file_path || !info) {
        set_last_error(L"Invalid parameters for updating placeholder");
        return -1;
    }

    CF_FS_METADATA fs_metadata = {};
    fs_metadata.BasicInfo.FileSize.QuadPart = info->file_size;
    fs_metadata.BasicInfo.FileAttributes = info->file_attributes;
    fs_metadata.BasicInfo.CreationTime.QuadPart = info->created_time;
    fs_metadata.BasicInfo.LastWriteTime.QuadPart = info->modified_time;
    fs_metadata.BasicInfo.LastAccessTime.QuadPart = info->modified_time;
    fs_metadata.BasicInfo.ChangeTime.QuadPart = info->modified_time;

    HRESULT hr = CfUpdatePlaceholder(
        file_path,
        &fs_metadata,
        &info->file_id,
        sizeof(info->file_id),
        nullptr,
        0,
        CF_UPDATE_FLAG_MARK_IN_SYNC,
        nullptr,
        nullptr
    );

    if (FAILED(hr)) {
        set_last_error(L"Failed to update placeholder. Error code: " + std::to_wstring(hr));
        return -2;
    }

    return 0;
}

// Get sync root status
int cf_get_sync_root_status(const wchar_t* sync_root_path) {
    if (!sync_root_path) return -1;

    CF_SYNC_ROOT_STANDARD_INFO info;
    HRESULT hr = CfGetSyncRootInfoByPath(
        sync_root_path,
        CF_SYNC_ROOT_INFO_STANDARD,
        &info,
        sizeof(info)
    );

    return SUCCEEDED(hr) ? 1 : 0;
}

// Set in-sync state for a file
int cf_set_in_sync_state(const wchar_t* file_path, int in_sync) {
    if (!file_path) return -1;

    DWORD flags = in_sync ? CF_SET_IN_SYNC_FLAG_NONE : CF_SET_IN_SYNC_FLAG_CLEAR;
    HRESULT hr = CfSetInSyncState(file_path, flags, nullptr);
    
    return SUCCEEDED(hr) ? 0 : -1;
}

// Register callbacks
int cf_register_callbacks(CloudFileCallbacks* callbacks) {
    g_callbacks = callbacks;
    return 0;
}

// Get placeholder state
int cf_get_placeholder_state(const wchar_t* file_path) {
    if (!file_path) return -1;

    CF_PLACEHOLDER_STATE state;
    HRESULT hr = CfGetPlaceholderStateFromFileInfo(
        nullptr, // Will get file info internally
        CF_PLACEHOLDER_STATE_PLACEHOLDER |
        CF_PLACEHOLDER_STATE_SYNC_ROOT |
        CF_PLACEHOLDER_STATE_ESSENTIAL_PROP_PRESENT,
        &state
    );

    if (FAILED(hr)) return -1;

    if (state & CF_PLACEHOLDER_STATE_PLACEHOLDER) {
        if (state & CF_PLACEHOLDER_STATE_PARTIAL) {
            return PLACEHOLDER_HYDRATION_PARTIAL;
        } else if (state & CF_PLACEHOLDER_STATE_PARTIALLY_ON_DISK) {
            return PLACEHOLDER_HYDRATION_PARTIAL;
        } else {
            return PLACEHOLDER_HYDRATION_NOT_HYDRATED;
        }
    }

    return PLACEHOLDER_HYDRATION_FULL;
}

// Check if Cloud Files is supported
int cf_is_cloud_files_supported(void) {
    // Cloud Files API requires Windows 10 version 1809 or later
    OSVERSIONINFOEX osvi = {};
    osvi.dwOSVersionInfoSize = sizeof(OSVERSIONINFOEX);
    osvi.dwMajorVersion = 10;
    osvi.dwMinorVersion = 0;
    osvi.dwBuildNumber = 17763; // Windows 10 1809

    DWORDLONG condition_mask = 0;
    VER_SET_CONDITION(condition_mask, VER_MAJORVERSION, VER_GREATER_EQUAL);
    VER_SET_CONDITION(condition_mask, VER_MINORVERSION, VER_GREATER_EQUAL);
    VER_SET_CONDITION(condition_mask, VER_BUILDNUMBER, VER_GREATER_EQUAL);

    return VerifyVersionInfo(&osvi, VER_MAJORVERSION | VER_MINORVERSION | VER_BUILDNUMBER, condition_mask);
}

// Get last error message
wchar_t* cf_get_last_error_message(void) {
    if (g_last_error.empty()) return nullptr;
    
    size_t len = g_last_error.length() + 1;
    wchar_t* result = (wchar_t*)malloc(len * sizeof(wchar_t));
    if (result) {
        wcscpy_s(result, len, g_last_error.c_str());
    }
    return result;
}

// Free wide string
void cf_free_wstring(wchar_t* str) {
    if (str) {
        free(str);
    }
}

#else // Non-Windows platforms

// Stub implementations for non-Windows platforms
int cf_register_sync_root(const wchar_t* sync_root_path, const wchar_t* display_name, 
                         const wchar_t* server_uri) { return -1; }
int cf_unregister_sync_root(const wchar_t* sync_root_path) { return -1; }
int cf_create_placeholders(const wchar_t* sync_root_path, FilePlaceholder* placeholders, 
                          int count) { return -1; }
int cf_hydrate_placeholder(const wchar_t* file_path, HydrationPolicy policy) { return -1; }
int cf_dehydrate_placeholder(const wchar_t* file_path) { return -1; }
int cf_update_placeholder(const wchar_t* file_path, FilePlaceholder* info) { return -1; }
int cf_get_sync_root_status(const wchar_t* sync_root_path) { return 0; }
int cf_set_in_sync_state(const wchar_t* file_path, int in_sync) { return -1; }
int cf_register_callbacks(CloudFileCallbacks* callbacks) { return -1; }
int cf_get_placeholder_state(const wchar_t* file_path) { return -1; }
int cf_is_cloud_files_supported(void) { return 0; }
wchar_t* cf_get_last_error_message(void) { return nullptr; }
void cf_free_wstring(wchar_t* str) {}

#endif