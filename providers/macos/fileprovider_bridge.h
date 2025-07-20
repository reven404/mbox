#ifndef FILEPROVIDER_BRIDGE_H
#define FILEPROVIDER_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

// File Provider Domain Management
typedef struct {
    char* domain_identifier;
    char* display_name;
    char* server_url;
    int enabled;
} FileProviderDomain;

// File Provider Operations
int fp_register_domain(const char* domain_id, const char* display_name, const char* server_url);
int fp_remove_domain(const char* domain_id);
int fp_signal_enumeration_change(const char* domain_id, const char* container_id);
int fp_signal_item_change(const char* domain_id, const char* item_id);

// Configuration Management
int fp_set_server_url(const char* domain_id, const char* server_url);
char* fp_get_server_url(const char* domain_id);
int fp_set_domain_enabled(const char* domain_id, int enabled);

// Status and Diagnostics
int fp_get_domain_status(const char* domain_id);
char* fp_get_domain_error(const char* domain_id);

// Memory management
void fp_free_string(char* str);

#ifdef __cplusplus
}
#endif

#endif // FILEPROVIDER_BRIDGE_H